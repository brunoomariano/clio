STATE:
  sidebar_selected: int = 0
  active_screen: string = "Dashboard"
  logs_filter: string = ""
  sidebar_collapsed: bool = false

EVENTS:

on_key("j"):
  sidebar_selected += 1

on_key("k"):
  sidebar_selected -= 1

on_key("enter"):
  active_screen = SIDEBAR.items[sidebar_selected]
