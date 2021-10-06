Eventually I'd like to add container scanning to the flow.

```yaml
scan:
  needs:
    - "dependencies"
    - "test"
    - "build"
    - "pg-test"
    - "container"
  name: Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - name: Download artifact
      uses: actions/download-artifact@v2
      with:
        name: '${{ github.sha }}-container-tar'
        path: '${{ github.workspace }}'
    - name: Load image
      run: |
        docker load --input ${{ github.workspace }}/rest-api.container.tar
        docker image ls -a
    # This one seems to need a bit more to provide meaningful output. Not something I'm going to
    # really work on right now.
    - uses: anchore/scan-action@v3
      with:
        image: 'containers.monetr.dev/rest-api:${{ github.sha }}'
        fail-build: false
    # I cannot currently use this action due to licensing limitations.
    # See: https://github.com/Azure/container-scan/issues/99
    - uses: azure/container-scan@v0
      name: Scan Image
      continue-on-error: true
      with:
        image-name: 'containers.monetr.dev/rest-api:${{ github.sha }}'
```


Caching built container image.

```yaml
- uses: actions/upload-artifact@v2
  with:
    name: '${{ github.sha }}-container-tar'
    path: '${{ github.workspace }}/rest-api.container.tar'
    retention-days: 7
```
