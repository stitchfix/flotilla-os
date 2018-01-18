FROM node:carbon
WORKDIR /usr/src/app
ADD . /usr/src/app
RUN npm install -g serve
RUN npm install
RUN npm rebuild node-sass
RUN npm run build
ENTRYPOINT serve -s build
