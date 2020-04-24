#reset env variables
unset role
unset workspace
unset helm_chart
unset operator
unset kind
unset api_version
#set env variable for export tool
export role="nginx"
export workspace="./workspace"
export helm_chart="./examples/helmcharts/nginx"

#Required for hack/build-operator
#export quay_namespace="YOUR_NAMESPACE"
#export INSTALL_OPERATOR_SDK=0

#Operator
#override this if required, else default is build in hack/init.sh
#export kind=                        #Version of the CR to be created.
#export api_version=                 #Kind of the CR to be created
#export operator=                    #operator name