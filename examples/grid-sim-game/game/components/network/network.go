// Package network holds the electrical circuit model for the grid-sim-game.
// It is a plain-Go graph resource (ElectricalNetwork) that lives alongside
// the ECS world but owns its own buses and branches. Grid entities are joined
// to the network via the NetworkLink component and the network's internal
// entityBus map.
package network

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// BusType distinguishes how a bus participates in power flow.
type BusType int

const (
	BusGenerator BusType = iota // slack / PV — power source
	BusLoad                     // PQ — power consumer (house)
	BusJunction                 // passive node (line segment)
)

// BusID uniquely identifies a bus within an ElectricalNetwork.
type BusID uint64

// BranchID uniquely identifies a branch within an ElectricalNetwork.
type BranchID uint64

// Bus is a node in the electrical network. It carries a back-link to the ECS
// entity it represents on the grid. Electrical quantities (voltage magnitude
// and angle, active/reactive power injection) are natural additions here when
// power-flow is implemented.
type Bus struct {
	ID     BusID
	Type   BusType
	Entity ecs.Entity
}

// Branch is an undirected connection between two buses representing a grid
// adjacency (line segment, or direct generator/house contact). Electrical
// parameters (resistance, reactance, thermal limit) go here when needed.
type Branch struct {
	ID   BranchID
	From BusID
	To   BusID
}

// NetworkLink is the ECS component that is the grid-side half of the join.
// Any entity that participates in the electrical network carries this so that
// ECS filters (e.g. Filter2[GridObject, NetworkLink]) can reach it without
// going through the network resource.
type NetworkLink struct {
	BusID BusID
}

// ElectricalNetwork is a per-World resource representing the power grid graph.
// It is the single source of truth for network topology and operating state.
// ECS entities hold only a NetworkLink back-reference.
type ElectricalNetwork struct {
	buses      map[BusID]*Bus
	branches   map[BranchID]*Branch
	entityBus  map[ecs.Entity]BusID // join: grid entity → bus
	incidence  map[BusID][]BranchID // adjacency list for fast Neighbors()
	nextBus    BusID
	nextBranch BranchID

	// State holds the current operating-point snapshot. It is kept in sync
	// with the topology: entries are created/removed with buses and branches.
	State *StaticState
}

// NewElectricalNetwork creates an empty network with an initialised StaticState.
func NewElectricalNetwork() *ElectricalNetwork {
	return &ElectricalNetwork{
		buses:     make(map[BusID]*Bus),
		branches:  make(map[BranchID]*Branch),
		entityBus: make(map[ecs.Entity]BusID),
		incidence: make(map[BusID][]BranchID),
		State:     newStaticState(),
	}
}

// --- Join interface ---------------------------------------------------------

// AddBus registers a new bus for the given ECS entity and returns it.
// Panics if the entity already has a bus.
func (n *ElectricalNetwork) AddBus(e ecs.Entity, t BusType) *Bus {
	if _, exists := n.entityBus[e]; exists {
		panic("network: entity already has a bus")
	}
	id := n.nextBus
	n.nextBus++
	b := &Bus{ID: id, Type: t, Entity: e}
	n.buses[id] = b
	n.entityBus[e] = id
	n.incidence[id] = nil
	n.State.Buses[id] = &BusState{Spec: defaultBusSpec(t)}
	return b
}

// BusForEntity returns the bus linked to the given ECS entity, or false if the
// entity has no bus (e.g. it is a blank tile).
func (n *ElectricalNetwork) BusForEntity(e ecs.Entity) (*Bus, bool) {
	id, ok := n.entityBus[e]
	if !ok {
		return nil, false
	}
	return n.buses[id], true
}

// Bus returns a bus by ID.
func (n *ElectricalNetwork) Bus(id BusID) (*Bus, bool) {
	b, ok := n.buses[id]
	return b, ok
}

// RemoveBus removes a bus and all of its incident branches from the network.
func (n *ElectricalNetwork) RemoveBus(id BusID) {
	for _, brID := range n.incidence[id] {
		br, ok := n.branches[brID]
		if !ok {
			continue
		}
		other := br.To
		if other == id {
			other = br.From
		}
		n.incidence[other] = removeBranchID(n.incidence[other], brID)
		delete(n.branches, brID)
	}
	if b, ok := n.buses[id]; ok {
		delete(n.entityBus, b.Entity)
	}
	delete(n.buses, id)
	delete(n.incidence, id)
	delete(n.State.Buses, id)
}

// --- Graph operations -------------------------------------------------------

// AddBranch connects two buses and returns the new branch.
// The connection is undirected: both buses' incidence lists are updated.
func (n *ElectricalNetwork) AddBranch(from, to BusID) *Branch {
	id := n.nextBranch
	n.nextBranch++
	br := &Branch{ID: id, From: from, To: to}
	n.branches[id] = br
	n.incidence[from] = append(n.incidence[from], id)
	n.incidence[to] = append(n.incidence[to], id)
	n.State.Branches[id] = &BranchState{}
	return br
}

// RemoveBranch removes a branch from the network.
func (n *ElectricalNetwork) RemoveBranch(id BranchID) {
	br, ok := n.branches[id]
	if !ok {
		return
	}
	n.incidence[br.From] = removeBranchID(n.incidence[br.From], id)
	n.incidence[br.To] = removeBranchID(n.incidence[br.To], id)
	delete(n.branches, id)
	delete(n.State.Branches, id)
}

// Neighbors returns all buses directly connected to the given bus.
func (n *ElectricalNetwork) Neighbors(id BusID) []*Bus {
	var result []*Bus
	for _, brID := range n.incidence[id] {
		br := n.branches[brID]
		other := br.To
		if other == id {
			other = br.From
		}
		if b, ok := n.buses[other]; ok {
			result = append(result, b)
		}
	}
	return result
}

func removeBranchID(s []BranchID, id BranchID) []BranchID {
	for i, v := range s {
		if v == id {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// --- State API (entity → solver boundary conditions, solver → caller results) ---

// SetBusSpec sets the boundary conditions for a bus. Called by any system
// that reads entity components and needs to inject known quantities into the
// solver (e.g. a generator entity sets its terminal voltage, a load entity
// sets its P/Q demand). It is a no-op if the bus ID is unknown.
func (n *ElectricalNetwork) SetBusSpec(id BusID, spec BusSpec) {
	if bs, ok := n.State.Buses[id]; ok {
		bs.Spec = spec
	}
}

// BusStateFor returns the full state (Spec + Result) for a bus, or false if
// the bus ID is unknown.
func (n *ElectricalNetwork) BusStateFor(id BusID) (*BusState, bool) {
	bs, ok := n.State.Buses[id]
	return bs, ok
}

// BranchStateFor returns the full state (Result) for a branch, or false if
// the branch ID is unknown.
func (n *ElectricalNetwork) BranchStateFor(id BranchID) (*BranchState, bool) {
	br, ok := n.State.Branches[id]
	return br, ok
}

// --- Read access for simulation systems ------------------------------------

// Buses returns the full bus map. Callers must not mutate the returned map.
func (n *ElectricalNetwork) Buses() map[BusID]*Bus {
	return n.buses
}

// Branches returns the full branch map. Callers must not mutate the returned map.
func (n *ElectricalNetwork) Branches() map[BranchID]*Branch {
	return n.branches
}

// Print prints a human-readable summary of all buses and branches to stdout,
// sorted by ID so the output is stable.
func (n *ElectricalNetwork) Print() {
	busIDs := make([]int, 0, len(n.buses))
	for id := range n.buses {
		busIDs = append(busIDs, int(id))
	}
	sort.Ints(busIDs)

	fmt.Printf("ElectricalNetwork: %d bus(es), %d branch(es)\n", len(n.buses), len(n.branches))
	fmt.Println("Buses:")
	for _, raw := range busIDs {
		b := n.buses[BusID(raw)]
		fmt.Printf("  bus %d  type=%-10s entity=%v\n", b.ID, b.Type, b.Entity)
	}

	branchIDs := make([]int, 0, len(n.branches))
	for id := range n.branches {
		branchIDs = append(branchIDs, int(id))
	}
	sort.Ints(branchIDs)

	fmt.Println("Branches:")
	for _, raw := range branchIDs {
		br := n.branches[BranchID(raw)]
		fmt.Printf("  branch %d  %d — %d\n", br.ID, br.From, br.To)
	}
}

// Log writes the same summary as Print but routes it through the engine
// logger at INFO level so it appears alongside other game log output.
func (n *ElectricalNetwork) Log() {
	busIDs := make([]int, 0, len(n.buses))
	for id := range n.buses {
		busIDs = append(busIDs, int(id))
	}
	sort.Ints(busIDs)

	branchIDs := make([]int, 0, len(n.branches))
	for id := range n.branches {
		branchIDs = append(branchIDs, int(id))
	}
	sort.Ints(branchIDs)

	var sb strings.Builder
	fmt.Fprintf(&sb, "network: %d bus(es), %d branch(es) |", len(n.buses), len(n.branches))
	for _, raw := range busIDs {
		b := n.buses[BusID(raw)]
		fmt.Fprintf(&sb, " bus%d(%s)", b.ID, b.Type)
	}
	sb.WriteString(" |")
	for _, raw := range branchIDs {
		br := n.branches[BranchID(raw)]
		fmt.Fprintf(&sb, " %d-%d", br.From, br.To)
	}
	logger.Logger.Info(sb.String())
}

func (t BusType) String() string {
	switch t {
	case BusGenerator:
		return "Generator"
	case BusLoad:
		return "Load"
	case BusJunction:
		return "Junction"
	default:
		return fmt.Sprintf("BusType(%d)", int(t))
	}
}
