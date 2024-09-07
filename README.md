# CommitFinder

# ⚠️ This repository has been archived because the method is outdated. I propose a new, simpler method [here](https://github.com/math-x-io/Removed-commits-finder):

CommitFinder is an OSINT tool that exploits vulnerabilities in GitHub commits. For more information, [see here](https://trufflesecurity.com/blog/anyone-can-access-deleted-and-private-repo-data-github).

The tool brute-forces the URL of a repository with random strings. The number of possible combinations is **60,466,176**. The time required to calculate each combination depends on your GPU. Generally, with a standard GPU, the duration is **around 60,466 secondes (~16 hours)**.

The tool is capable of retrieving **deleted** or **private** commits.


**Build command**
```sh
go build
```




