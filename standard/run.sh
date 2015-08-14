#!/bin/bash
COMMAND="gcloud preview app run ./app.yaml"
echo Running: $COMMAND
$COMMAND
