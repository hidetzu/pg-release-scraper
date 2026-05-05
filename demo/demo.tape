# vhs script for demo.gif
# Regenerate with:  vhs demo/demo.tape
#
# Prereqs:
#   - vhs installed (brew install vhs)
#   - bin/pg-release-scraper built (make build)
#   - network access to https://www.postgresql.org

Output demo/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1100
Set Height 650
Set Theme "Dracula"
Set TypingSpeed 50ms
Set PlaybackSpeed 1.0

Hide
Type "export PATH=$PWD/bin:$PATH && rm -rf /tmp/demo-out && clear"
Enter
Sleep 500ms
Show

Type "pg-release-scraper --version"
Enter
Sleep 800ms

Type "# 1) baseline: no filtering"
Enter
Type "pg-release-scraper --start 14.5 --end 14.6 --format md --output /tmp/demo-out"
Enter
Sleep 13s

Type "ls /tmp/demo-out/"
Enter
Sleep 1500ms

Hide
Type "clear"
Enter
Sleep 300ms
Show

Type "# 2) with rules: drop low-impact items, keep them in Excel marked as 対象外"
Enter
Type "pg-release-scraper --start 14.5 --end 14.6 --rules examples/rules/app-impact.yaml --format md --output /tmp/demo-out"
Enter
Sleep 13s

Type "ls /tmp/demo-out/"
Enter
Sleep 2500ms
