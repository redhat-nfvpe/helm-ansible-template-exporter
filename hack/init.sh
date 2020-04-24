#!/usr/bin/env bash
#Environament varaiables
export role=${role:-"nginx"}
export workspace=${workspace:-"./workspace"}
export helm_chart=${helm_chart:-"./examples/helmcharts/${role}/"}
#Change this to your namespace
export quay_namespace=${quay_namespace:-"YOUR_NAMESPACE"}
export INSTALL_OPERATOR_SDK=${INSTALL_OPERATOR_SDK:-0}

#Required for the operator-sdk to build operator
export kind=${kind:-"$(tr '[:lower:]' '[:upper:]' <<<${role:0:1})${role:1}"} #Version of the CR to be created.
export api_version=${api_version:-"app.${role}.com/v1alpha1"}                #Kind of the CR to be created
export operator=${operator:-"${role}-operator"}                              #operator name


