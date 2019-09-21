# Flotilla UI

The Flotilla UI is a React application bundled along with the rest of Flotilla. If you are running the entire Flotilla stack locally, it is recommended to use docker-compose as documented in the main [README](https://github.com/stitchfix/flotilla-os#starting-the-service-locally). If you are interested in developing the UI itself, you can follow these steps:

### Development
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

#### Develop
Note: when developing or building the UI, you will need to add a `REACT_APP_BASE_URL` environment variable as shown below. Additionally, you can pass a `REACT_APP_DEFAULT_CLUSTER` environment variable to autocomplete the cluster when launching a task (this is optional).

```
REACT_APP_BASE_URL="http://flotilla-api.com/api"      REACT_APP_DEFAULT_CLUSTER="my-flotilla-cluster" npm start
```

#### Test
UI testing is done with Jest and Enzyme. You can run the tests via:
```
npm run test
```

#### Build
While it is recommended to serve the UI as part of the entire Flotilla stack, you can build a production version of the UI via:
```
REACT_APP_BASE_URL="http://flotilla-api.com/api"  REACT_APP_DEFAULT_CLUSTER="my-flotilla-cluster" npm run build
```
