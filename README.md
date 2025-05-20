# k8s_requirementsCheck-crane
 Requirements check code for crane on Kubernetes

### ENV variables we need

- WORKING_NAMESPACE
- ROLE_NAME
- ROLE_BINDING_NAME
- SERVICE_ACCOUNT_NAME
- SV_ENABLE


## Image

- Alpine based basic image
- Should have binary for crane-testhook
- Bin should be executable
- Should be executed non-root
- Image should be publicly accessible. 

