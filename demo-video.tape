# Tarra Claw — Demo Video
Output docs/demo.mp4

Set FontSize 15
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 30
Set TypingSpeed 60ms
Set PlaybackSpeed 1.0

# ─── SCENE 1: The problem ──────────────────────────────────────
# Show a file with bugs before review

Type "# Scene 1: You've written some code..."
Enter
Sleep 1s

Type "cat /tmp/claw-demo/main.go"
Enter
Sleep 3s

Type "# Looks fine right? Let's ask Tarra Claw..."
Enter
Sleep 1.5s

# ─── SCENE 2: claw review ──────────────────────────────────────

Type "cd /tmp/claw-demo && TARRA_PROVIDER=ollama claw review --staged"
Enter
Sleep 40s

# ─── SCENE 3: claw commit ──────────────────────────────────────

Type "# Now let's write the commit message automatically"
Enter
Sleep 1s

Type "TARRA_PROVIDER=ollama claw commit"
Enter
Sleep 35s

# ─── SCENE 4: interactive chat ─────────────────────────────────

Type "# And the full interactive AI chat..."
Enter
Sleep 1s

Type "claw --help"
Enter
Sleep 4s

Sleep 2s
