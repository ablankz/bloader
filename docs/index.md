<!-- ---
layout: default
title: "Welcome to Bloader"
nav_order: 1
---

# Bloader Documentation 🚀

Welcome to the official documentation for **Bloader**, the modern benchmarking tool that simplifies load testing. Whether you're a seasoned developer or just starting, Bloader provides flexibility and power for all your testing needs.

### 🛠 Features
- Internal Store for managing requests.
- **Master-Slave Architecture** with gRPC for distributed testing.
- YAML-based configuration with **Sprig** template engine.

---



<div align="center">
  <a href="https://github.com/ablankz/bloader" class="btn btn-primary">View on GitHub</a>
</div> -->


---
title: Home
layout: home
nav_order: 1
description: "Documentation for Bloader, the modern benchmarking tool that simplifies load testing. Whether you're a seasoned developer or just starting, Bloader provides flexibility and power for all your testing needs"
permalink: /
---

# Bloader Documentation 🚀
{: .fs-9 }

Bloader is a benchmark testing project that focuses on flexibility and simplicity.
{: .fs-6 .fw-300 }

[Get started now](#getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View it on GitHub][Bloader repo]{: .btn .fs-5 .mb-4 .mb-md-0 }

---

{: .warning }
> This website documents the features of the current `main` branch of Bloader. See [the CHANGELOG]({% link CHANGELOG.md %}) for a list of releases, new features, and bug fixes.

Welcome to the official documentation for **Bloader**, the modern benchmarking tool that simplifies load testing. Whether you're a seasoned developer or just starting, Bloader provides flexibility and power for all your testing needs.

### 🛠 Features
- Internal Store for managing requests.
- **Master-Slave Architecture** with gRPC for distributed testing.
- YAML-based configuration with **Sprig** template engine.

Browse the docs to learn more about how to use bloader.

### 📖 Reference Software
- [Sprig](https://masterminds.github.io/sprig/): Template engine 
- [Cobra](https://github.com/spf13/cobra): CLI creation 
- [Viper](https://github.com/spf13/viper): Parsing Config files
- [Buf](https://buf.build/): gRPC schema management 
- [Bolt](https://github.com/boltdb/bolt): Internal store 

## Getting started

<!-- The [Bloader Template] provides the simplest, quickest, and easiest way to create a new website that uses the Bloader theme. To get started with creating a site, just click "[use the template]"! -->

{: .note }
To use the theme, you do ***not*** need to clone or fork the [Bloader repo]! You should do that only if you intend to browse the theme docs locally, contribute to the development of the theme, or develop a new theme based on Bloader.

You can easily set the site created by the template to be published on [GitHub Pages] – the [template README] file explains how to do that, along with other details.

<!-- If [Jekyll] is installed on your computer, you can also build and preview the created site *locally*. This lets you test changes before committing them, and avoids waiting for GitHub Pages.[^2] And you will be able to deploy your local build to a different platform than GitHub Pages. -->

More specifically, the created site:

<!-- - uses a gem-based approach, i.e. uses a `Gemfile` and loads the `just-the-docs` gem
- uses the [GitHub Pages / Actions workflow] to build and publish the site on GitHub Pages -->

Other than that, you're free to customize sites that you create with the template, however you like. You can easily change the versions of `just-the-docs` and Jekyll it uses, as well as adding further plugins.

{: .note }
See [README][Bloader README] for a brief documentation.

## Future implementations 
- Change from BoltDB to a database that is still supported today
- Add external cloud providers, etc. to Override's Type.
- Add functionality for performing analysis.
- Make it run as a server and make it visually clear besides the CLI.
- Introduce original Encrypt between Master and Slave.
- Add gRPC as a measurement target.
- Addition of test code 
- Add plugin functionality. 

## About the project

Bloader is &copy; 2024-{{ "now" | date: "%Y" }} by [Hayashi Kenta](k.hayashi@cresplanex.com).

### License

Bloader is distributed by an [MIT license](https://github.com/ablankz/bloader/tree/main/LICENSE).

### Contributing

When contributing to this repository, please first discuss the change you wish to make via issue,
email, or any other method with the owners of this repository before making a change. Read more about becoming a contributor in [our GitHub repo](https://github.com/ablankz/bloader#contributing).

#### Thank you to the contributors of Bloader!

<ul class="list-style-none">
{% for contributor in site.github.contributors %}
  <li class="d-inline-block mr-1">
     <a href="{{ contributor.html_url }}"><img src="{{ contributor.avatar_url }}" width="32" height="32" alt="{{ contributor.login }}"></a>
  </li>
{% endfor %}
</ul>

### Code of Conduct

Bloader is committed to fostering a welcoming community.

[View our Code of Conduct](https://github.com/ablankz/bloader/tree/main/CODE_OF_CONDUCT.md) on our GitHub repository.

[Bloader]: https://docs.bloader.cresplanex.org.com
[Bloader repo]: https://github.com/ablankz/bloader
[Bloader README]: https://github.com/ablankz/bloader/blob/main/README.md
[GitHub Pages]: https://pages.github.com/

