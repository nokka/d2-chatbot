# D2-Chatbot
D2 Chatbot provides a global chat for any Diablo 2 server running on PvpGN and provides in game community chat channels.

## Package dependency graph
![Package dependency graph](docs/deps.png)

### How to generate the graph
[godephgraph](https://github.com/kisielk/godepgraph) is used to generate the dependency graph.

```bash
$ godepgraph -nostdlib -novendor -horizontal -onlyprefixes=github.com/nokka/d2-chatbot github.com/nokka/d2-chatbot/cmd/server | dot -Tpng -o docs/deps.png
```
---
