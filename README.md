# Yum Mirror Checker Utility

This shim verifies that a repository is complete and has no missing or invalid files.  Any file
that is invalid or missing will be printed to the output as a filelist, ready for downloading
via the filelist-mirror-downloader utility.

# Example usage:
```bash
./yum-check-mirror -insecure -path test/ -repo "/7/os/x86_64" -debug
```

# Usage help:
```bash
$ ./yum-check-mirror -h
Yum Check Mirror,  Version: 0.1.20220412.2212

Usage: ./yum-check-mirror [options...]

  -debug
        Turn on debug, more verbose
  -insecure
        Skip signature checks
  -keyring string
        Use keyring for verifying, keyring.gpg or keys/ directory (default "keys/")
  -output string
        Output file to put the results of the check (default "-")
  -path string
        Path to the mirror base (default ".")
  -repo string
        Repo to check (default "/")
```
