COMPONENT Sidebar:
  props:
    - items: string[]
    - selected: int
    - collapsed: bool

COMPONENT Panel:
  props:
    - title: string
    - border: bool
    - height: int

COMPONENT Table:
  props:
    - columns: string[]
    - rows: string[][]
