# Flotilla UI
The UI for Flotilla is a React/Redux application built on top of [`create-react-app`](https://github.com/facebookincubator/create-react-app) and [`custom-react-scripts`](https://github.com/kitze/custom-react-scripts).

## Quickstart
To get started, we've packaged all the necessary steps to run the UI on a local development server in a [`Dockerfile`](./Dockerfile). Run the following command in the `/ui` directory to start the UI server at <http://localhost:49160/>.

```sh
docker run \
-e REACT_APP_FLOTILLA_API_ROOT="http://my-flotilla-api.com/api/v1" \
-e REACT_APP_CLUSTERS_API_ROOT="http://my-clusters.com/api/v1" \
-e REACT_APP_IMAGE_ENDPOINT="http://my-docker-images.com/api/v1" \
-e REACT_APP_IMAGE_TAGS_ENDPOINT="http://my-docker-image-tags/{image}/some-other-stuff" \
-e REACT_APP_DOCKER_REPOSITORY_HOST="my-docker-repository:5000" \
-p 49160:3000 -d psun/flotilla-ui
```

## Building the UI
- Install Node (and NPM) on your machine. Directions for doing so can be found [here](https://docs.npmjs.com/getting-started/installing-node).
- Run the following commands in the `/ui` directory:

```sh
# Install JS modules.
npm install
# Build the UI with the necessary environment variables.
# The resulting build files will be written to the `/ui/build` dir.
REACT_APP_FLOTILLA_API_ROOT="http://my-flotilla-api.com/api/v1" REACT_APP_CLUSTERS_API_ROOT="http://my-clusters.com/api/v1" REACT_APP_IMAGE_ENDPOINT="http://my-docker-images.com/api/v1" REACT_APP_IMAGE_TAGS_ENDPOINT="http://my-docker-image-tags/{image}/some-other-stuff" REACT_APP_DOCKER_REPOSITORY_HOST="my-docker-repository:4321" npm run build
```

## Environment Variables
The following environment variables are required for the UI to work.
- `REACT_APP_FLOTILLA_API_ROOT`
- `REACT_APP_CLUSTERS_API_ROOT`
- `REACT_APP_IMAGE_ENDPOINT`
- `REACT_APP_IMAGE_TAGS_ENDPOINT`
- `REACT_APP_DOCKER_REPOSITORY_HOST`
