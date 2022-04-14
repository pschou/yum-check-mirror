# Yum Mirror Checker Utility

This shim verifies that a repository is complete and has no missing or invalid files.  Any file
that is invalid or missing will be printed to the output as a filelist, ready for downloading
via the filelist-mirror-downloader utility.

# Example usage:
```bash
./yum-check-mirror -insecure -path test/ -repo "/7/os/x86_64" -debug -output missing.txt
```

```bash
$ cat missing.txt
{sha256}845e42288d3b73a069e781b4307caba890fc168327baba20ce2d78a7507eb2af 1623852 repodata/845e42288d3b73a069e781b4307caba890fc168327baba20ce2d78a7507eb2af-other.xml.gz
{sha256}e595924b51a69153c2148f0f4b3fc2c31a1ad3114a6784687520673740e4f54a 289524 Packages/389-ds-base-devel-1.3.10.2-6.el7.x86_64.rpm
{sha256}4fbe999bb69d21a8b753cfad4e543413166fe3222ec742e56193494d430adba6 182748 Packages/389-ds-base-snmp-1.3.10.2-6.el7.x86_64.rpm
{sha256}037a51e8aeed759dcf683e360694a8313747371b4dacebbc1c34631bfda4f6ff 2027512 Packages/Cython-0.19-5.el7.x86_64.rpm
{sha256}d03cfa43fe3fc14ca5fbb7cf1c0d636c4e59e342a69537668bf7eeb6fcb614bc 35772 Packages/ElectricFence-2.2.2-39.el7.i686.rpm
{sha256}716b1bf12d8921d143eb05c5252bbd24d5d8e50be7cbf762b9b0530e05a4f519 36116 Packages/ElectricFence-2.2.2-39.el7.x86_64.rpm
{sha256}d0df638cf1bb17be8a1ef4105144a4dc67d462e7ba092314de0add664f3e1908 1046312 Packages/GConf2-3.2.6-8.el7.i686.rpm
{sha256}3d4f93baccf4e3bf657e013b91d5695cb92ff661810cddb26e560f224531b5fd 1047864 Packages/GConf2-3.2.6-8.el7.x86_64.rpm
{sha256}20cfcdfd6b86b385bebe50d2031436b06f3d8f9943f1b72f08c66fa2d283aa38 112696 Packages/GConf2-devel-3.2.6-8.el7.i686.rpm
{sha256}a3439f8e58a3800ed6d1769daaa4580f5a7732e00c38311cf64ea9da24f2656b 112656 Packages/GConf2-devel-3.2.6-8.el7.x86_64.rpm
```

## Prune old packages:
Here we create an empty file which is not found in the repo metadata and run prune to verify and then delete the file:
```
# create an empty rpm
$ touch test/Packages/my_null.rpm

# verify the list to be deleted
$ ./yum-check-mirror -insecure -path test/ -repo / -prune-test -output files_to_get.txt -multi
Scanning for files to delete in test/
-  test/Packages/my_null.rpm

# to delete
$ ./yum-check-mirror -insecure -path test/ -repo / -prune -output files_to_get.txt -multi
```

# Usage help:
```bash
$ ./yum-check-mirror -h
Yum Check Mirror,  Version: 0.1.20220413.2222

Usage: ./yum-check-mirror [options...]

  -debug
        Turn on debug, more verbose
  -insecure
        Skip signature checks
  -keyring string
        Use keyring for verifying, keyring.gpg or keys/ directory (default "keys/")
  -multi
        Scan for multiple package lists in repo directory.  Note: Secondary lists are
        insecure as they are missing the GPG signature file, and may not be a complete set!
  -output string
        Output file to put the file list, the failed results of the check (default "-")
  -path string
        Path to the mirror base (default ".")
  -prune
        Find and remove un-used packages in repo (.rpm)
  -prune-test
        Find and display all un-used packages in repo (.rpm)
  -repo string
        Repo to check (example "/7/os/x86_64")
  -repodata string
        Explicit path to /repodata/ to check (example "/downloads/yum/20220101/")
```
