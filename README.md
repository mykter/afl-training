# Fuzzing with AFL workshop

Materials of the "Fuzzing with AFL" workshop by Michael Macnair (@michael_macnair).

This workshop introduces fuzzing and how to make the most of using American Fuzzy Lop, a popular and powerful fuzzer,
through a series of challenges where you rediscover real vulnerabilities in popular open source projects.

The first public version of this workshop was presented at SteelCon 2017 and it was revised for each of BSides London
2019, BSides Bristol 2019, and GrayHat 2020 (most notable change in this revision was a switch to
[afl++](https://github.com/AFLplusplus/AFLplusplus)).

# Presentation

Via
[Google slides](https://docs.google.com/presentation/d/1Ap3eUIo4RrI_9GAGfn2Q0RKtBPNmGhI1DZlivdcFKPo)
and [as a PDF](https://github.com/mykter/afl-training/files/5454345/Fuzzing.with.AFL.-.GrayHat.2020.pdf). There is extra
information in the speaker notes.

GrayHat published [a recording of a remote version of the workshop](https://www.youtube.com/watch?v=6YLz9IGAGLw) on
YouTube - this was created for a real-time workshop audience, but you can follow along at your own pace as long as you
don't mind skipping a few pauses and ignoring references to Discord.

The presentation suggests when to attempt the different challenges in this repository, and the video provides a
walk-through of `quickstart` and `harness`.

# Pre-requisites

- 3-4 hours (more to complete all the challenges)
- Linux machine
- Basic C and command line experience - ability to modify and compile C programs.
- Docker, or the dependencies described in `quickstart`.

# Contents

- quickstart - Do this first! A tiny sample program to get started with fuzzing, including instructions on how to setup
  your machine.
- harness - the basics of creating a test harness. Do this if you have any doubts about the "plumbing" between afl-fuzz
  and the target code.
- challenges - a set of known-vulnerable programs with fuzzing hints
- docker - Instructions and Dockerfile for preparing a suitable environment, and hosting it on GCP if you wish. A
  prebuilt image can be pulled from [ghcr.io/mykter/fuzz-training](https://ghcr.io/mykter/fuzz-training).

See the other READMEs for more information.

# Challenges

Challenges, roughly in recommended order, with any specific aspects they cover:

- libxml2 - an ideal target, using ASAN and persistent mode.
- heartbleed - infamous bug, using ASAN.
- sendmail/1301 - parallel fuzzing
- ntpq - fuzzing a network client; coverage analysis and increasing coverage
- date - fuzzing environment variable input
- cyber-grand-challenge - an easy vuln and an example of a hard to find vuln using afl
- sendmail/1305 - persistent mode difficulties

The challenges have HINTS.md and ANSWERS.md files - these contain useful information about fuzzing different targets
even if you're not going to attempt the challenge.

Most of the challenges also have an ANSWERS-libFuzzer.md file, for if you want to try out using LLVM's libFuzzer. These
are brief descriptions of the differences for libFuzzer, and should be read alongside the afl docs (.md files).

All of the challenges use real vulnerabilities from open source projects (the CVEs are identified in the descriptions),
with the exception of the Cyber Grand Challenge extract, which is a synthetic vulnerability.

The chosen bugs are all fairly well isolated, and (except where noted) are very amenable to fuzzing. This means that you
should be able to discover the bugs with a relatively small amount of compute time - these won't take core-days, most of
them will take core-minutes. That said, fuzz testing is by definition a random process, so there's no guarantee how long
it will take to find a particular bug, just a probability distribution.

# Links

- The afl [docs/](https://github.com/AFLplusplus/AFLplusplus/tree/stable/docs) directory
- Ben Nagy’s “Finding Bugs in OS X using AFL” [video](https://vimeo.com/129701495)
- The [afl-users mailing list](https://groups.google.com/forum/#!forum/afl-users)
- The smart fuzzer revolution (talk on the future of fuzzing): [video](https://www.youtube.com/watch?v=g1E2Ce5cBhI) /
  [slides](https://docs.google.com/presentation/d/1FgcMRv_pwgOh1yL5y4GFsl1ozFwd6PMNGlMi2ONkGec/edit#slide=id.g13a9c1bce4_6_0)
- A [categorized collection of recent fuzzing papers](https://github.com/wcventure/FuzzingPaper) (there are a lot!)
- [The Fuzzing Book](https://www.fuzzingbook.org/) - broad coverage of fuzzing
- [libFuzzer](http://llvm.org/docs/LibFuzzer.html)
  - [libFuzzer workshop](https://github.com/Dor1s/libfuzzer-workshop)
  - [libFuzzer tutorial](https://github.com/google/fuzzer-test-suite/blob/master/tutorial/libFuzzerTutorial.md)
- [More challenges](https://github.com/antonio-morales/EkoParty_Advanced_Fuzzing_Workshop) from an EkoParty workshop
- Introduction to [triaging crashes](https://trustfoundry.net/introduction-to-triaging-fuzzer-generated-crashes/)
- Google's [ClusterFuzz](https://github.com/google/clusterfuzz) and Microsoft's
  [OneFuzz](https://github.com/microsoft/onefuzz)
