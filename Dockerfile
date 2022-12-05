#  This file is part of the eliona project.
#  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
#  ______ _ _
# |  ____| (_)
# | |__  | |_  ___  _ __   __ _
# |  __| | | |/ _ \| '_ \ / _` |
# | |____| | | (_) | | | | (_| |
# |______|_|_|\___/|_| |_|\__,_|
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
#  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
#  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
#  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

FROM golang:1.18-alpine3.15 AS build

WORKDIR /
COPY . ./

RUN go mod download
RUN go build -o ../bitcoiner

FROM alpine:3.15 AS target

COPY --from=build /bitcoiner ./
COPY conf/*.sql ./conf/
COPY conf/*.yaml ./conf/

ENV APPNAME=bitcoiner
ENV CONNECTION_STRING=postgres://postgres:secret@database-mock:5432
ENV API_ENDPOINT=http://api-v2:3000/v2
ENV API_TOKEN=secret
ENV API_SERVER_PORT=8082
ENV APPNAME=bitcoiner
ENV DEBUG_LEVEL=info

ENV TZ=Europe/Zurich
CMD [ "/bitcoiner" ]
