FROM node:16
WORKDIR /usr/src/app
ADD . /usr/src/app
RUN npm install -g serve typescript
RUN npm install --legacy-peer-deps

ARG FLOTILLA_API
ARG DEFAULT_CLUSTER

ENV REACT_APP_FLOTILLA_API=${FLOTILLA_API}
ENV REACT_APP_DEFAULT_CLUSTER=${DEFAULT_CLUSTER}
ENV NODE_OPTIONS=--max-old-space-size=4096
ENV TSC_COMPILE_ON_ERROR=true
ENV GENERATE_SOURCEMAP=false
ENV CI=false

RUN npm run build || true
EXPOSE 5000
ENTRYPOINT serve -s build -l 5000