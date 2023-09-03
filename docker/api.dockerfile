# node 
FROM node:20 AS builder

RUN apt-get update && apt-get install -y npm

WORKDIR /app
COPY ./webapp/js/package.json /app/package.json
RUN yarn install

FROM node:20 AS isuda
WORKDIR /app
COPY ./webapp/js /app
COPY ./webapp/public /app/public
COPY --from=builder /app/node_modules /app/node_modules

CMD 'yarn' 'isuda'

FROM node:20 AS isutar
WORKDIR /app
COPY ./webapp/js /app
COPY ./webapp/public /app/public
COPY --from=builder /app/node_modules /app/node_modules

CMD 'yarn' 'isutar'
