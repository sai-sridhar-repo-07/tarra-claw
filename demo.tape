# Tarra Claw — real demo
Output docs/demo.gif

Set FontSize 13
Set Width 1100
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 70ms

Type "cd /tmp/claw-demo"
Enter
Sleep 400ms

Type "git diff --staged --stat"
Enter
Sleep 2s

Type "TARRA_PROVIDER=ollama claw review --staged"
Enter
Sleep 35s
