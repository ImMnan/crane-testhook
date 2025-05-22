# k8s_requirementsCheck-crane
 Requirements check code for crane on Kubernetes

### ENV variables

- WORKING_NAMESPACE
- ROLE_NAME
- ROLE_BINDING_NAME
- SERVICE_ACCOUNT_NAME
- SV_ENABLE


## Image

- Alpine based basic image
- binary for crane-testhook
- executed non-root
- Image  publicly accessible. 


## compile 

 go env -w GOOS=linux GOARCH=amd64 && go build -o crane-testhook
