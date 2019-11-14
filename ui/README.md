# Flotilla UI

The Flotilla UI is a React application bundled along with the rest of Flotilla. If you are running the entire Flotilla stack locally, it is recommended to use docker-compose as documented in the main [README](https://github.com/stitchfix/flotilla-os#starting-the-service-locally). If you are interested in developing the UI itself, you can follow these steps:

## Development

### Running Locally

```
git clone git@github.com:stitchfix/flotilla-os.git
cd flotilla-os/ui
npm install
REACT_APP_BASE_URL=http://my-flotilla.com REACT_APP_BASE_URL_DEV=http://localhost:5000/api npm start
```

### Testing

UI testing is done with Jest and Enzyme. You can run the tests via:

```
npm run test
```
