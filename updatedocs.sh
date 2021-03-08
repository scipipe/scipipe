#!/bin/bash
if [[ ! -d site ]]; then
    git clone git@github.com:scipipe/scipipe.github.io.git site
fi && \
echo "Building docs site..." && \
mkdocs build && \
echo "Moving into site folder..." && \
cd site/ && \
echo "Recreating CNAME file..." && \
echo "scipipe.org" > CNAME && \
echo "Adding new files to git ..." && \
git add * && \
echo "Committing changes in site repo..." && \
git commit -am "Update docs" && \
echo "Pushing to docs site" && \
git push && \
echo "Moving back to main folder..." && \
cd ../ && \
echo "Done."
