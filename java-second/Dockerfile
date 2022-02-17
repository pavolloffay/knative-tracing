FROM eclipse-temurin:8 as builder

ENV APP_HOME /app/
WORKDIR $APP_HOME

COPY . .
RUN ./mvnw package -Dlicense.skip=true -DskipTests && rm -rf ~/.m2

FROM eclipse-temurin:8-jre
MAINTAINER Pavol Loffay <ploffay@redhat.com>
ENV APP_HOME /app/
COPY --from=builder $APP_HOME/target/cloudevents-spring-reactive-example-0.1.0-exec.jar $APP_HOME/

WORKDIR $APP_HOME

ADD https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v1.11.0/opentelemetry-javaagent.jar /javaagent.jar

CMD java -javaagent:/javaagent.jar ${JAVA_OPTS} -jar $APP_HOME/cloudevents-spring-reactive-example-0.1.0-exec.jar
