#!/usr/bin/env bash
#
# Generate a changelog entry for a proposed PR by grabbing the next available
# auto incrementing ID in GitHub.

if ! command -v curl &> /dev/null
then
  echo "jq not be found"
  exit 1
fi

if ! command -v jq &> /dev/null
then
  echo "jq not be found"
  exit 1
fi

current_pr=$(curl -s "https://api.github.com/repos/cloudflare/cloudflare-go/issues?state=all&per_page=1" | jq -r ".[].number")
next_pr=$(($current_pr + 1))
changelog_path=".changelog/$next_pr.txt"

echo "==> What type of change is this? (enhancement, bug, breaking-change)"
read entry_type

echo "==> What is the summary of this change? Example: dns: updated X to do Y"
read entry_summary

touch $(pwd)/$changelog_path
cat > $(pwd)/$changelog_path <<EOF
\`\`\`release-note:$entry_type
$entry_summary
\`\`\`
EOF

echo
echo "Successfully created $changelog_path. Don't forget to commit it and open the PR!"
