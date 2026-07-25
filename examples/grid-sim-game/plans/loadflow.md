# Load Flow & Network Analysis Architecture

## Layers

```
ElectricalNetwork  (topology + state container)
│
├── StaticState    (one solved snapshot: bus voltages, branch flows)
│   ├── map[BusID]*BusState
│   └── map[BranchID]*BranchState
│
├── Solver interface
│   └── LoadflowSolver   (stub → Newton-Raphson / DC approx later)
│
└── (future) TimeEvolution
    ├── []StaticState    (series of snapshots)
    └── Step(dt)         (advance simulation by one time step)
```

---

## State types (`network/state.go`)

### BusSpec — boundary conditions (entity → solver)

What the grid entity (or a system reading its ECS components) provides
before solving. Meaningful fields depend on BusType:

| BusType   | Fixed fields          | Free fields       |
|-----------|-----------------------|-------------------|
| Generator (swing) | VoltMag, VoltAng | PInject, QInject |
| Generator (PV)    | VoltMag, PInject  | VoltAng, QInject |
| Load (PQ)         | PInject, QInject  | VoltMag, VoltAng |
| Junction (PQ=0)   | PInject=0, QInject=0 | VoltMag, VoltAng |

```go
type BusSpec struct {
    VoltMag float64  // |V| per-unit (swing/PV buses)
    VoltAng float64  // voltage angle radians (swing bus only)
    PInject float64  // active power injection (+ve = generation)
    QInject float64  // reactive power injection
}
```

### BusResult — solver output

```go
type BusResult struct {
    VoltMag float64
    VoltAng float64
    PInject float64
    QInject float64
}
```

### BranchResult — solver output

```go
type BranchResult struct {
    CurrentMag float64  // |I| per-unit
    PFrom      float64  // active power flow from→to
    PTo        float64  // active power flow to→from
    QFrom      float64  // reactive flow from→to
    QTo        float64  // reactive flow to→from
}
```

### StaticState — the snapshot container

```go
type StaticState struct {
    Buses      map[BusID]*BusState
    Branches   map[BranchID]*BranchState
    Converged  bool
    Iterations int
}
```

Entries are created/destroyed automatically in `AddBus` / `RemoveBus` /
`AddBranch` / `RemoveBranch` — StaticState always mirrors the topology.

---

## Entity → Solver API

The ECS system (e.g. a future `NetworkSetupSystem`) reads entity components
and calls these before every solve:

```go
net.SetBusSpec(busID, BusSpec{VoltMag: 1.0, VoltAng: 0})  // swing
net.SetBusSpec(busID, BusSpec{PInject: -0.5, QInject: -0.1}) // load
```

After solving, results are read back:

```go
bs, _ := net.BusStateFor(busID)    // → *BusState {Spec, Result}
br, _ := net.BranchStateFor(brID)  // → *BranchState {Result}
```

These could drive visual feedback (e.g. colour tiles by voltage level,
highlight overloaded branches).

---

## Solver interface (`network/solver.go`)

```go
type Solver interface {
    Solve(net *ElectricalNetwork) error
}
```

### LoadflowSolver (stub)

Flat-start initialisation: sets all bus voltages to 1.0∠0° and all branch
flows to zero. Marks the state `Converged=true` with `Iterations=0`.

Real implementations will replace this with:
- **DC load flow** — linear, fast, ignores reactive power (good for topology checks)
- **Newton-Raphson AC** — full nonlinear solve for V, θ, P, Q

---

## Time evolution

Not implemented. House demand currently re-samples via `LoadTickSystem`;
history rings (`BusHistory` / `BranchHistory`) cover recent solve samples.
A dedicated TimeEvolution stepper can be added later if schedules need
explicit snapshots beyond Dirty-driven solves.

---

## File layout

```
game/components/network/
  network.go   ← topology graph + State field + entity-join API
  state.go     ← BusSpec, BusResult, BusState, BranchResult, BranchState, StaticState
  solver.go    ← Solver interface, LoadflowSolver
  history.go   ← Series / BusHistory / BranchHistory
  ybus.go      ← Y-bus build + CalcPQ
  sparse.go    ← SparseMatrix
```
