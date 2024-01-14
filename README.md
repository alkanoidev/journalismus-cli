# Journalismus
> [!NOTE]
> Capture thoughts effortlessly in the command line.

https://github.com/charmbracelet/bubbletea/blob/master/examples/list-simple/main.go

### Roadmap:
- store in markdown
- custom dir input
- config
    - mood rating
    - insert date in the top of entry
    - templating system
    - store by month (1 month = 1 md file)
- view
  - use list to render md files
    - read all .md from directory
    - group by months/years
  - filename cmd
  - sort entries
  - filter
- after init show prompt to run view or write
- root cmd if init shows prompt to run view or write
- apply fullscreen to all cmds
- fix: journal view today not working on install
- handle write error
- create event for file selection
- put glamour element inside viewport