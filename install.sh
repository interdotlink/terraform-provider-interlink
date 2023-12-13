#!/bin/env bash
DOMAIN=inter.link
ORG=tech
PLUGIN=interlink
VERSION=0.1.0

PLATFORM=$(terraform -version -json |jq -r .platform)
INSTALL_DIR=${HOME}/.terraform.d/plugins/${DOMAIN}/${ORG}/${PLUGIN}/${VERSION}/${PLATFORM}
go build -o ${INSTALL_DIR}/terraform-provider-${PLUGIN}_v${VERSION}
