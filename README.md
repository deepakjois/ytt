As of Jun 2025, I have stopped using this. It's too much effort to keep up with the way it breaks everytime Google makes some changes to limit the use of this API.

I now use the upstream [youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api) directly (it is actively maintained) with [uv](https://docs.astral.sh/uv/) inside a bash function like this:

```bash
# Usage: yttn <youtube_video_url>
function yttn {
uvx --with youtube-transcript-api python -c "
import sys
import re
from youtube_transcript_api import YouTubeTranscriptApi
from youtube_transcript_api.formatters import TextFormatter

if len(sys.argv) < 2:
    print('Usage: provide YouTube URL or video_id as argument', file=sys.stderr)
    sys.exit(1)

url_or_id = sys.argv[1]

# Extract video ID from URL if it's a full URL
video_id_match = re.search(r'(?:v=|/)([a-zA-Z0-9_-]{11})', url_or_id)
if video_id_match:
    video_id = video_id_match.group(1)
else:
    # Assume it's already a video ID
    video_id = url_or_id

try:
    ytt_api = YouTubeTranscriptApi()
    transcript = ytt_api.fetch(video_id)
    formatter = TextFormatter()
    print(formatter.format_transcript(transcript))
except Exception as e:
    print(f'Error: {e}', file=sys.stderr)
    sys.exit(1)
" $1 | pbcopy
}
```


# ytt
[![Go Reference](https://pkg.go.dev/badge/github.com/deepakjois/ytt.svg)](https://pkg.go.dev/github.com/deepakjois/ytt)

Fetch YouTube transcripts. Adapted from [youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api).

### Install

```
go install github.com/deepakjois/ytt/cmd/ytt@latest
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

