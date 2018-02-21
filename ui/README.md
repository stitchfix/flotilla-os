# Flotilla UI

The Flotilla UI is a React application bundled along with the rest of Flotilla. If you are running the entire Flotilla stack locally, it is recommended to use docker-compose as documented in the main [README](https://github.com/stitchfix/flotilla-os#starting-the-service-locally). If you are interested in developing the UI itself, you can follow these steps:

### Development
#### Prerequsites
- [Node 8](https://nodejs.org/en/)
- [NPM 5](https://www.npmjs.com)

#### Setup
1. Clone the repo

```
git clone git@github.com:stitchfix/flotilla-os.git
cd flotilla-os/ui
```

2. Ensure that your local version of Node is 8.x and NPM is 5.x. It is highly recommended to use [nvm](https://github.com/creationix/nvm) to manage your Node versions locally.

```
node -v
# Should output 8.x.x
npm -v
# Should output 5.x.x

# If you are using nvm, you can run the following command to ensure that your Node version is correct:
nvm use
```

3. Install dependencies

```
npm install
```

##### Develop
1. Start Webpack development server
Note: when developing or building the UI, you will need to add a `FLOTILLA_API` environment variable as shown below. Additionally, you can pass a `DEFAULT_CLUSTER` environment variable to autocomplete the cluster when launching a task (this is optional).

```
FLOTILLA_API="http://flotilla-api.com/api/v1"  DEFAULT_CLUSTER="my-flotilla-cluster" npm start
```

2. Go to [locahost:8080](locahost:8080)

##### Test
UI testing is done with Jest and Enzyme. You can run the tests once via `npm run test` or have Jest watch for changes via:

```
npm run test:watch
```

##### Build
While it is recommended to serve the UI as part of the entire Flotilla stack, you can build a production version of the UI via:

```
FLOTILLA_API="http://flotilla-api.com/api/v1"  DEFAULT_CLUSTER="my-flotilla-cluster" npm run build
```
