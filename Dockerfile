FROM       python:3.8-slim-buster
ENV        format="sql"
RUN        apt-get update \
                && apt-get install -y curl apt-transport-https gnupg2 \
                && curl https://packages.microsoft.com/keys/microsoft.asc | apt-key add - \
                && curl https://packages.microsoft.com/config/debian/10/prod.list > /etc/apt/sources.list.d/mssql-release.list \
                && apt-get update \
                && ACCEPT_EULA=Y apt-get install -y msodbcsql17 \
                && ACCEPT_EULA=Y apt-get install -y mssql-tools
RUN        apt-get install -y \
           build-essential \
           libffi6 \
           libffi-dev \
           git \
           libssl-dev \
           libxml2-dev \
           libxslt-dev \
           unixodbc \
           unixodbc-dev
RUN        git clone https://github.com/SafetyCulture/iauditor-exporter /app
WORKDIR    /app
RUN        pip install -r requirements.txt
ENTRYPOINT python exporter.py --docker --loop --format $format
