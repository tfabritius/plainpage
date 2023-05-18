#/bin/sh

# Fail on errors
set -e

# Change directory to location of this script
cd "$(dirname "$0")"

cd ../frontend
echo "Building frontend..."
pnpm install --frozen-lockfile
pnpm generate
echo "Building fontend done."

cd ../backend
go generate ./...
echo "Building backend..."
GOOS=darwin GOARCH=amd64 go build -o ../bin/plainpage-amd64-darwin .
GOOS=darwin GOARCH=arm64 go build -o ../bin/plainpage-arm64-darwin .
GOOS=linux GOARCH=amd64 go build -o ../bin/plainpage-amd64-linux .
GOOS=linux GOARCH=arm64 go build -o ../bin/plainpage-arm64-linux .
GOOS=windows GOARCH=amd64 go build -o ../bin/plainpage-windows-amd64.exe .
echo "Building backend done."

echo "Uploading release assets..."
if [ -z "$GITHUB_TOKEN" ] || [ -z "$GITHUB_REPO" ] || [ -z "$GITHUB_RELEASE" ]; then
  echo "Error: GITHUB_TOKEN, GITHUB_REPO, and GITHUB_RELEASE environment variables must be set."
  exit 1
fi

for file in ../bin/*; do
  echo "Uploading $file..."
  curl \
    --header "Authorization: token $GITHUB_TOKEN" \
    --header "Content-Type: application/octet-stream" \
    --data-binary @"$file" \
    --url "https://uploads.github.com/repos/$GITHUB_REPO/releases/$GITHUB_RELEASE/assets?name=$(basename $file)"
done

echo "Uploading done."
