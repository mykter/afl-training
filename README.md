# Fuzzing with AFL workshop

Materials of the "Fuzzing with AFL" workshop by Michael Macnair (@michael_macnair).

The first public version of this workshop was presented at SteelCon 2017 and it was revised for BSides London and Bristol 2019.

# Pre-requisites

- 3-4 hours (more to complete all the challenges)
- Linux machine
- Basic C and command line experience - ability to modify and compile C programs.
- Docker, or the dependencies described in `quickstart`.

# Contents

- quickstart - Do this first! A tiny sample program to get started with fuzzing, including instructions on how to setup your machine.
- harness - the basics of creating a test harness. Do this if you have any doubts about the "plumbing" between afl-fuzz and the target code.
- challenges - a set of known-vulnerable programs with fuzzing hints
- docker - Instructions and Dockerfile for preparing a suitable environment, and hosting it on AWS if you wish. A prebuilt image is on Docker Hub at [mykter/afl-training](https://hub.docker.com/r/mykter/afl-training).

See the other READMEs for more information.

# Challenges

Challenges, roughly in recommended order, with any specific aspects they cover:

- libxml2 - an ideal target, using ASAN and persistent mode.
- heartbleed - infamous bug, using ASAN.
- sendmail/1301 - parallel fuzzing
- date - fuzzing environment variable input
- ntpq - fuzzing a network client; coverage analysis and increasing coverage
- cyber-grand-challenge - an easy vuln and an example of a hard to find vuln using afl
- sendmail/1305 - persistent mode difficulties

The challenges have HINTS.md and ANSWERS.md files - these contain useful information about fuzzing different targets even if you're not going to attempt the challenge.

All of the challenges use real vulnerabilities from open source projects (the CVEs are identified in the descriptions), with the exception of the Cyber Grand Challenge extract, which is a synthetic vulnerability.

The chosen bugs are all fairly well isolated, and (except where noted) are very amenable to fuzzing. This means that you should be able to discover the bugs with a relatively small amount of compute time - these won't take core-days, most of them will take core-minutes. That said, fuzz testing is be definition a random process, so there's no guarantee how long it will take to find a particular bug, just a probability distribution.

# Slides

Via [Google slides](https://drive.google.com/file/d/1g78GgmMtxn_Aei2L1K6UzaCQmjaqiUNj/view?usp=sharing) and in [PowerPoint format](https://github.com/mykter/afl-training/files/3325293/Fuzzing.with.AFL.pptx). There is extra information in the speaker notes.

# Links

- The afl docs/ directory
- Ben Nagy’s “Finding Bugs in OS X using AFL” [video](https://vimeo.com/129701495)
- The [afl-users mailing list](https://groups.google.com/forum/#!forum/afl-users)
- The smart fuzzer revolution (talk on the future of fuzzing): [video](https://www.youtube.com/watch?v=g1E2Ce5cBhI) / [slides](https://docs.google.com/presentation/d/1FgcMRv_pwgOh1yL5y4GFsl1ozFwd6PMNGlMi2ONkGec/edit#slide=id.g13a9c1bce4_6_0)
- A [categorized collection of recent fuzzing papers](https://github.com/wcventure/FuzzingPaper) (there are a lot!)
- [The Fuzzing Book](https://www.fuzzingbook.org/) - broad coverage of fuzzing
- [libFuzzer](http://llvm.org/docs/LibFuzzer.html)
  - [libFuzzer workshop](https://github.com/Dor1s/libfuzzer-workshop)
  - [libFuzzer tutorial](https://github.com/google/fuzzer-test-suite/blob/master/tutorial/libFuzzerTutorial.md)
- [More challenges](https://github.com/antonio-morales/EkoParty_Advanced_Fuzzing_Workshop) from an EkoParty workshop
- Introduction to [triaging crashes](https://trustfoundry.net/introduction-to-triaging-fuzzer-generated-crashes/)
- Google's [ClusterFuzz](https://github.com/google/clusterfuzz) and Microsoft's [OneFuzz](https://github.com/microsoft/onefuzz)
