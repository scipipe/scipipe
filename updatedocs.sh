#!/bin/bash
read -p "Please write a commit message for the doc change: " msg
echo "Committing changes in main repo ..."
git commit -am "$msg"
echo "Building docs site..."
mkdocs build
echo "Moving into site folder..."
cd site/
echo "Recreating CNAME file..."
echo "scipipe.org" > CNAME
echo "Adding new files to git ..."
git add *
echo "Committing changes in site repo..."
git commit -am "$msg"
echo "Pushing to docs site"
git push
echo "Moving back to main folder..."
cd ../
echo "Done."
