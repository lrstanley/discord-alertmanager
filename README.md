<!-- template:define:options
{
  "nodescription": true
}
-->
![logo](https://liam.sh/-/gh/svg/lrstanley/discord-alertmanager?icon=logos%3Aprometheus&icon.height=80&bg=topography&layout=left)

<!-- template:begin:header -->
<!-- do not edit anything in this "template" block, its auto-generated -->

<p align="center">
  <a href="https://github.com/lrstanley/discord-alertmanager/tags">
    <img title="Latest Semver Tag" src="https://img.shields.io/github/v/tag/lrstanley/discord-alertmanager?style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/discord-alertmanager/commits/master">
    <img title="Last commit" src="https://img.shields.io/github/last-commit/lrstanley/discord-alertmanager?style=flat-square">
  </a>




  <a href="https://github.com/lrstanley/discord-alertmanager/actions?query=workflow%3Atest+event%3Apush">
    <img title="GitHub Workflow Status (test @ master)" src="https://img.shields.io/github/actions/workflow/status/lrstanley/discord-alertmanager/test.yml?branch=master&label=test&style=flat-square">
  </a>

  <a href="https://codecov.io/gh/lrstanley/discord-alertmanager">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/lrstanley/discord-alertmanager/master?style=flat-square">
  </a>

  <a href="https://pkg.go.dev/github.com/lrstanley/discord-alertmanager">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/lrstanley/discord-alertmanager?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/lrstanley/discord-alertmanager">
    <img title="Go Report Card" src="https://goreportcard.com/badge/github.com/lrstanley/discord-alertmanager?style=flat-square">
  </a>
</p>
<p align="center">
  <a href="https://github.com/lrstanley/discord-alertmanager/issues?q=is:open+is:issue+label:bug">
    <img title="Bug reports" src="https://img.shields.io/github/issues/lrstanley/discord-alertmanager/bug?label=issues&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/discord-alertmanager/issues?q=is:open+is:issue+label:enhancement">
    <img title="Feature requests" src="https://img.shields.io/github/issues/lrstanley/discord-alertmanager/enhancement?label=feature%20requests&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/discord-alertmanager/pulls">
    <img title="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lrstanley/discord-alertmanager?label=prs&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/discord-alertmanager/discussions/new?category=q-a">
    <img title="Ask a Question" src="https://img.shields.io/badge/support-ask_a_question!-blue?style=flat-square">
  </a>
  <a href="https://liam.sh/chat"><img src="https://img.shields.io/badge/discord-bytecord-blue.svg?style=flat-square" title="Discord Chat"></a>
</p>
<!-- template:end:header -->

<!-- template:begin:toc -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :link: Table of Contents

  - [Why](#grey_question-why)
  - [Installation](#computer-installation)
    - [Container Images (ghcr)](#whale-container-images-ghcr)
    - [Source](#toolbox-source)
  - [Usage](#gear-usage)
    - [Slash Commands](#green_book-slash-commands)
    - [Message Commands](#speech_balloon-message-commands)
  - [Support &amp; Assistance](#raising_hand_man-support--assistance)
  - [Contributing](#handshake-contributing)
  - [License](#balance_scale-license)
<!-- template:end:toc -->

## :grey_question: Why

I utilize the direct integration between **AlertManager** and **Discord**, via webhooks. I also
wanted a way of managing my silences directly from the same channels where those alerts were
being posted, thus why I created this bot. The bot allows you to both manage silences directly
via Discords [slash commands](https://support.discord.com/hc/en-us/articles/1500000368501-Slash-Commands-FAQ),
as well as interact with AlertManagers webhook messages, to silence specific alerts,
pre-configuring the silence with content from that specific alert.

## :computer: Installation

Check out the [releases](https://github.com/users/lrstanley/discord-alertmanager/pkgs/container/discord-alertmanager)
page for prebuilt versions.

<!-- template:begin:ghcr -->
<!-- do not edit anything in this "template" block, its auto-generated -->
### :whale: Container Images (ghcr)

```console
$ docker run -it --rm ghcr.io/lrstanley/discord-alertmanager:1.0.0
$ docker run -it --rm ghcr.io/lrstanley/discord-alertmanager:latest
$ docker run -it --rm ghcr.io/lrstanley/discord-alertmanager:master
```
<!-- template:end:ghcr -->

### :toolbox: Source

Note that you must have [Go](https://golang.org/doc/install) installed (latest is usually best).

    git clone https://github.com/lrstanley/discord-alertmanager.git && cd discord-alertmanager
    make
    ./discord-alertmanager --help

## :gear: Usage

Take a look at the [docker-compose.yaml](/docker-compose.yaml) file, or the above
docker run commands. For references on supported flags/environment variables,
take a look at [USAGE.md](/USAGE.md)

### :green_book: Slash Commands

You can utilize Discords [slash commands](https://support.discord.com/hc/en-us/articles/1500000368501-Slash-Commands-FAQ),
and `add`, `get`, `edit`, `list`, and `remove` silences:

Full list of commands:

![slash commands](https://cdn.liam.sh/share/2023/06/Discord_lwhrUPJClx.png)

Example for fetching a specific silence:

![/silences get id](https://cdn.liam.sh/share/2023/06/Discord_AjvCN7Pe4b.gif)

Example for listing all active silences:

![/silences list](https://cdn.liam.sh/share/2023/06/Discord_yjcapcwsMp.gif)

### :speech_balloon: Message Commands

You can right click AlertManager webhook events, and add a silence:

![silence alert from webhook](https://cdn.liam.sh/share/2023/06/Discord_9zJVqHDfvg.gif)

You can right click a silence, and update it:

![edit silence](https://cdn.liam.sh/share/2023/06/Discord_oZx5NCaeWA.gif)

Or remove it:

![remove silence](https://cdn.liam.sh/share/2023/06/Discord_oqByuoYFUI.gif)

<!-- template:begin:support -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :raising_hand_man: Support & Assistance

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for
     guidelines on ensuring everyone has the best experience interacting with
     the community.
* :raising_hand_man: Take a look at the [support](.github/SUPPORT.md) document on
     guidelines for tips on how to ask the right questions.
* :lady_beetle: For all features/bugs/issues/questions/etc, [head over here](https://github.com/lrstanley/discord-alertmanager/issues/new/choose).
<!-- template:end:support -->

<!-- template:begin:contributing -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :handshake: Contributing

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for guidelines
     on ensuring everyone has the best experience interacting with the
    community.
* :clipboard: Please review the [contributing](.github/CONTRIBUTING.md) doc for submitting
     issues/a guide on submitting pull requests and helping out.
* :old_key: For anything security related, please review this repositories [security policy](https://github.com/lrstanley/discord-alertmanager/security/policy).
<!-- template:end:contributing -->

<!-- template:begin:license -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :balance_scale: License

```
MIT License

Copyright (c) 2023 Liam Stanley <me@liamstanley.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

_Also located [here](LICENSE)_
<!-- template:end:license -->
