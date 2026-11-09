# The output shown is replayed from a real run on a kind cluster. To regenerate,
# from the repo root:
#   make build && examples/demo/capture.sh
#   docker run --rm -v "$PWD:/vhs" ghcr.io/charmbracelet/vhs examples/demo/demo.tape
Output examples/demo/demo.gif

Set Shell bash
Set FontSize 16
Set Width 1280
Set Height 680
Set Padding 16
Set TypingSpeed 35ms
Set Theme "Catppuccin Mocha"

Hide
Type "source examples/demo/replay.sh"
Enter
Sleep 500ms
Show

Type "kube-diff helm examples/helm/demo-chart/ -r demo -n kube-diff-demo"
Enter
Sleep 6s

Hide
Type "clear"
Enter
Sleep 300ms
Show
Type "kube-diff helm examples/helm/demo-chart/ -r demo -n kube-diff-demo -l app=demo --deleted -o table"
Enter
Sleep 5s
