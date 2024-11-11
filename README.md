# VeraDemo - Blab-a-Gag (GOlang) 

### Notice
This project is intentionally vulnerable! It contains known vulnerabilities and security errors in its code and is meant as an example project for software security scanning tools such as Veracode. Please do not report vulnerabilities in this project; the odds are they’re there on purpose :) .

## About
Blab-a-Gag is a simple forum type application that allows:

- Users can post a one-liner joke.
- Users can follow the jokes of other users or not (listen or ignore).
- Users can comment on other users messages (heckle).

### URLs
- `/feed` shows the jokes/heckles that are relevant to the current user.
- `/blabbers` shows a list of all other users and allows the current user to listen or ignore.
- `/profile` allows the current user to modify their profile.
- `/login` allows you to log in to your account
- `/register` allows you to create a new user account
- `/tools` shows a tools page that shows a fortune or lets you ping a host.
- `/reset` allows the user to reset the database

## Run Docker Image
If you don't already have Docker this is a prerequisite.  Make sure your Docker engine is up to date.  To check, run:

	docker compose version

The version should be >= 2.22.

(NOTE: Currently only local development is supported - there are no images to download from Docker Hub)


### Download Docker
Visit [docker desktop](https://www.docker.com/products/docker-desktop/) and download your compatible version.  Follow installation instructions.  Open the Docker app.

To run the container, clone the respository and in the project folder run this:

    docker compose up -d

For development

Run:

    docker compose up -d
    docker compose watch

Navigate to [http://localhost:8000](http://localhost:8000).

Then register as a new user and add some feeds.

## Run Locally
To run locally, you will need to install the latest version of [Golang](https://go.dev/dl/).

To verify install, open command line/terminal and run:

    go

This should print all the go commands that are available, verifying that install was sucessful. 

Run the application:

    cd verademo-go
    go run app.go

Navigate to [http://localhost:8000](http://localhost:8000).

Then register as a new user and add some feeds.

## Exploitations/Flaws

See the [DEMO_NOTES](DEMO_NOTES.md) file for information on using this application with the various Veracode scan types.

Also see the `docs` folder for in-depth explanations of the various exploits exposed in this application.

## Technologies Used

- Golang (v1.22.5)
- [TOTP Library for GO by pquerna](https://github.com/pquerna/otp) 
