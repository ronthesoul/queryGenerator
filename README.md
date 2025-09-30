# QueryGen -- Security Search Query Builder

Generate large boolean search queries like:

    "VENDOR_01 hacked" ~10 OR "VENDOR_02 cve" ~10 OR ...

from a simple text file of vendors and keywords. Output is saved to a
`.txt` file you can paste into your search platform.

------------------------------------------------------------------------

## Features

-   Reads vendors and keywords from a text file (supports labels, CSV
    lines, or two-block format).
-   Builds `"vendor keyword" ~<proximity>` clauses joined with `OR`.
-   Optional tail filter (e.g., `NOT crawler: paste*`).
-   GitHub Action lints, builds cross-platform binaries, auto-tags each
    push, and publishes releases.

------------------------------------------------------------------------

## Install

### Option A: Download a release

1.  Go to **Releases** on this repo.

2.  Download the binary for your OS/arch (e.g.,
    `querygen-darwin-arm64`).

3.  Make it executable and move to PATH:

    ``` bash
    chmod +x ./querygen-darwin-arm64
    mv ./querygen-darwin-arm64 /usr/local/bin/querygen
    ```

### Option B: Build from source

``` bash
go build -o querygen .
```

------------------------------------------------------------------------

## Input file format

Create an `input.txt` (examples below). Comments start with `#`.

### Labeled CSV (recommended)

``` txt
# Vendors and keywords (placeholders recommended for public repos)
vendors:
  VENDOR_01, VENDOR_02, VENDOR_03, VENDOR_04

keywords:
  hacked, breached, compromised, exploited, 0day, cve, vulnerability
```

### Single-line CSV per section

``` txt
vendors: VENDOR_01, VENDOR_02, VENDOR_03
keywords: hacked, breached, 0day, cve
```

### Two-block (vendors first, keywords second)

``` txt
VENDOR_01, VENDOR_02, VENDOR_03

hacked, breached, compromised, cve
```

> Tip: Use **neutral placeholders** (e.g., `VENDOR_01`) in public repos.
> Keep a private mapping file locally.

------------------------------------------------------------------------

## Usage

### Interactive mode

``` bash
./querygen
# Enter path to .txt file (vendors + keywords): ./input.txt
# Enter proximity number (e.g., 10): 10
# Enter output filename (e.g., query.txt): query.txt
# Optional tail filter (e.g., NOT crawler: paste*). Leave empty for none: NOT crawler: paste*
```

------------------------------------------------------------------------

## Output

The tool writes a single line to the chosen output file, e.g.:

``` txt
"VENDOR_01 hacked" ~10 OR "VENDOR_01 cve" ~10 OR "VENDOR_02 hacked" ~10 OR ...
  ... plus your optional tail filter at the very end.
```

------------------------------------------------------------------------


## Example

1.  Create `input.txt`:

    ``` txt
    vendors:
      VENDOR_01, VENDOR_02, VENDOR_03

    keywords:
      hacked, breached, compromised, exploited, 0day, cve, vulnerability
    ```

2.  Generate:

    ``` bash
    ./querygen -in input.txt -prox 12 -out big-query.txt -tail 'NOT crawler: paste*'
    ```

3.  Use `big-query.txt` wherever you need the OR-joined query string.

------------------------------------------------------------------------


