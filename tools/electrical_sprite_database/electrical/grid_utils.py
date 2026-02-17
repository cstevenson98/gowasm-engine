"""
Utility functions for grid-based sprite generation.
"""

import math
from typing import List, Tuple


def enumerate_grid_offsets(radius: int, include_negative: bool = False) -> List[Tuple[int, int]]:
    """
    Generate all (x,y) grid offsets within a given radius.
    
    Args:
        radius: Maximum distance from origin
        include_negative: If True, include negative offsets (e.g., (-1, 0))
    
    Returns:
        List of (x, y) tuples within radius
    """
    offsets = []
    
    start = -radius if include_negative else 0
    
    for x in range(start, radius + 1):
        for y in range(start, radius + 1):
            # Skip origin
            if x == 0 and y == 0:
                continue
            
            # Check if within radius
            dist = math.sqrt(x * x + y * y)
            if dist <= radius:
                offsets.append((x, y))
    
    return offsets


def get_phase_index(phase: str) -> int:
    """
    Convert phase letter to connection point index.
    
    Args:
        phase: Phase letter ('A', 'B', or 'C')
    
    Returns:
        Connection point index (0, 1, or 2)
    """
    phase_map = {'A': 0, 'B': 1, 'C': 2}
    
    if phase.upper() not in phase_map:
        raise ValueError(f"Invalid phase '{phase}'. Must be 'A', 'B', or 'C'")
    
    return phase_map[phase.upper()]


def get_phase_letter(index: int) -> str:
    """
    Convert connection point index to phase letter.
    
    Args:
        index: Connection point index (0, 1, or 2)
    
    Returns:
        Phase letter ('A', 'B', or 'C')
    """
    phases = ['A', 'B', 'C']
    
    if index < 0 or index >= len(phases):
        raise ValueError(f"Invalid index {index}. Must be 0, 1, or 2")
    
    return phases[index]


def format_offset_name(offset_x: int, offset_y: int) -> str:
    """
    Format grid offset as string for filenames.
    
    Args:
        offset_x: X offset
        offset_y: Y offset
    
    Returns:
        Formatted string like "x1_y2" or "x-1_y0"
    """
    return f"x{offset_x}_y{offset_y}"


if __name__ == "__main__":
    # Test enumeration
    print("Grid offsets within radius 3:")
    offsets = enumerate_grid_offsets(3)
    print(f"Total offsets: {len(offsets)}")
    for offset in sorted(offsets):
        print(f"  {offset}")
    
    print("\nWith negative offsets:")
    offsets_full = enumerate_grid_offsets(3, include_negative=True)
    print(f"Total offsets: {len(offsets_full)}")


