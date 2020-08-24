# Raise Money For Partners In Health

For the past decade, the Vlogbrothers and Nerdfigtheria continuously support Partners In Health as a beneificiary of the annual [Project for Awesome](http://www.projectforawesome.com). In 2019, Hank and John started an inititive to fund a new maternal care facility in Sierra Leon to help reduce maternal morality in the Kona district, where Partner's In Health works tirelessly to improve healthcare.

## But first, some background

Read about the Green family connection to Partner's In Health on the [Partner's In Health website](https://www.pih.org/vlogbrothers-support-maternal-health). Find more about the [fundraiser for this materity center in this Vlogbrothers video on YouTube](https://www.youtube.com/watch?v=DwDjsNFHVhQ).

## This Bot

Allows Twitter users to raise money for Partner's in Health. If you see a tweet so good you want to donate to Partner's In Health on behalf of that tweeter, respond to the tweet and [@CharityYeti](https://twitter.com/charityyeti). We'll send you a link to donate to Partner's in Health and will tweet back letting the original tweeter that you appreciated their tweet to support maternal health in Sierra Leon! [This tweet by Hank](https://twitter.com/hankgreen/status/1186824079120011264) is the basis of this project. An image of the tweet is saved here as `goal.jpeg`.

## Deployment

I built Charity Yeti on [Docker](https://www.docker.com) using [docker-compose](https://docs.docker.com/compose/). You can spin up a backend by cloning the repo and issuing `docker-compose up charityyeti` from the same directory as the `Dockerfile` (the repo root). This exposes the port set as your environment variable `PORT`. See the dependecies section below for setting environment variables for local development easily.

## Code of Conduct

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v1.4%20adopted-ff69b4.svg)](code-of-conduct.md)

This project is by Nerdfighters, for Partners In Health. We would love open collaboration and accept contributions. We also expect that you not forget to be awesome, and adhere by the [Contributer Covenant](https://www.contributor-covenant.org). See `code-of-conduct.md` for full details.

## Contributing
This is an early stage project. To contribute, please open an issue so we can discuss. Once we address the issue at hand, fork the repo and make the necessary changes or updates, then create a pull request into this repo.

## Dependencies
I started building this with [GoDotEnv](https://github.com/joho/godotenv) and [env](https://github.com/caarlos0/env) to manage environment variables. I'm also using [Go-Twitter](https://github.com/dghubble/go-twitter#authentication) to do Twitter interactions.