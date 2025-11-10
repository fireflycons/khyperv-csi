# Hyper-V volume management service

Runs as a service on the Hyper-V host machine. Effectively all the packages within this directory structure amount to providing a "cloud provider"-like REST API to the controller service running in-cluster.

Performs the low-level operations to manage VHDs:

| Operation     | Description                                         | REST method | Sample                                                 |
|---------------|-----------------------------------------------------|-------------|--------------------------------------------------------|
| `Create`      | Provisions a VHD                                    | `POST`      | `http://backend/volume/:name?size=n`                   |
| `Delete`      | Deletes a VHD                                       | `DELETE`    | `http://backend/volume/:volid`                         |
| `Get`         | Gets a VHD                                          | `GET`       | `http://backend/volume/:volid`                         |
| `List`        | Lists available VHDs (with pagination)              | `GET`       | `http://backend/volumes?maxEntries=n&nextToken=n`      |
| `Expand`      | Expands a VHD                                       | `PUT`       | `http://backend/volume/:volId?size=n`                   |
| `Attach`      | Attach a VHD to a VM                                | `PUT`       | `http://backend/attachment/node/:nodeid/volume/:volid` |
| `Detach`      | Remove a VHD from a VM                              | `DELETE`    | `http://backend/attachment/node/:nodeid/volume/:volid` |
| `GetCapacity` | Return available storage space for VHDs on the host | `GET`       | `http://backend/capacity`                              |
| `ListVms`     | Return all VMs on the host                          | `GET`       | `http://backend/vms`                                   |
| `GetVm`       | Return a VM by ID                                   | `GET`       | `http://backend/vm/:id`                                |
