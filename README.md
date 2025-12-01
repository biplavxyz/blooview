# BlooView
Filesystem and Network Monitoring Tool For Linux that runs within a Terminal.

# Installation
- Install Golang
- Clone the repo
`git clone https://github.com/biplavxyz/blooview`
- Create blooview config path
`mkdir -p ~/.config/blooview; touch ~/.config/blooview/blooview.toml`
- Build and run
`go build -o ./bin/blooview . && ./bin/blooview`

# Progress
- [x] Display filesystem changes.
- [x] Display established network connection information.
- [ ] Display live commands being run by established net process.
- [ ] Write tests.
