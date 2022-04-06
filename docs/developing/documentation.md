# Documentation

This is an overview of ways to contribute to monetr's documentation. To get started:

- You can find outstanding issues for documentation
  here: [![GitHub issues by-label](https://img.shields.io/github/issues/monetr/monetr/documentation)](https://github.com/monetr/monetr/issues?q=is%3Aopen+is%3Aissue+label%3Adocumentation){:target="_blank"}
- If you don't find an issue that you'd be interested in working on, you can still create a pull request with your
  desired changes.
- If you have found a gap in our documentation that you aren't able to, or do not wish to fill yourself; please create
  an issue so that others are aware of this gap, and it can be addressed.

## Editing documentation

All of our documentation is in the form of Markdown files in the `docs` directory of the monetr repository. You can
simply edit the existing files to make changes to the documentation. The documentation site is automatically generated
in our GitHub Actions workflows.

??? note

    To properly work on documentation locally, it is recommended that you run the local development stack. Documentation
    for that can be found in [Developer / Local](local.md)

You can build our documentation site locally using the following command, but it does require a Docker runtime to be
available.

```shell title="Shell"
make mkdocs
```

Our documentation is built using the insider build of [mkdocs-material](https://github.com/squidfunk/mkdocs-material)
but can still be built locally using the normal version of mkdocs-material. If you do have access to an insider
container image, you can specify that image when you run the make command like this:

```shell title="Shell"
make mkdocs MKDOCS_IMAGE=ghcr.io/yourusername/mkdocs-material-insiders:yourTag
```

## Style

We would like our documentation to follow a general guide, this creates some consistency in how our documentation is
both written, presented, and maintained over time.

### Language

All documentation should be written in "American" English as much as possible. The exception to that rule are
quotations, trademarks or terms that are better known by their own language's equivalent.

### Reader / Author

The documentation prefers "we" to address the author and "you" to address the reader. The gender of the reader shall be
neutral if possible. Attempt to use "they" as a pronoun for the reader.

### Code Blocks

Code blocks should always be accompanied by a preceding text to give context as to what that code block is, or
represents. Adjacent code blocks without a paragraph of text between them should be avoided.

### Screenshots

Screenshots, if at all possible, should be no larger than `1280x720`. This is not a strict requirement, but if a
screenshot can reasonably capture all the necessary details in that resolution or less; that is greatly preferred.

### Links

Links to external sites should be opened in a new tab. This can be done by appending the following snippet after a link.

```text title="Open In New Tab"
{:target="_blank"}
```

### Inclusivity

Language that has been identified as hurtful or insensitive should be avoided.
