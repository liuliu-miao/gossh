# echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | sudo tee /etc/apt/sources.list.d/goreleaser.list
# sudo apt update
# sudo apt install goreleaser
rm -rf dist
goreleaser release --snapshot --skip-publish --rm-dist

