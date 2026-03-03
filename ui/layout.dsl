SCREEN: Dashboard
WIDTH: 80
HEIGHT: 24

LAYOUT: vertical

HEADER (height=3)
  title="Sistema"

BODY (flex=1, layout=horizontal)

  SIDEBAR (width=24)
    items=["Home","Logs","Config"]
    selected=0

  MAIN (flex=1, layout=vertical)
    PANEL (height=8, title="Resumo", border=true)
    TABLE (flex=1, columns=["id","status","tempo"])

FOOTER (height=1)
  hint="q=Sair | r=Refresh"
