# ytt
Fetch YouTube transcripts. Adapted from [youtube-transcript-api](https://github.com/jdegoes/youtube-transcript-api).

### Install

```
go install github.com/deepakjois/ytt@latest
```

Make sure `$HOME/go/bin` is in path.

### Usage

```
$ ytt -h
ytt <youtube_url>
  -lang string
        Language code for the desired transcript (default "en")
  -no-timestamps
        Don't print timestamps
  -o string
        Output filename (defaults to stdout)
```

### Library

```
import "github.com/deepakjois/ytt"
```

For detailed API documentation, visit [pkg.go.dev](https://pkg.go.dev/github.com/deepakjois/ytt).

