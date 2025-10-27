//go:build windows

package messages

const (
	SERVER_STARTING        = "server starting"
	SERVER_STOPPING        = "server stopping"
	SERVER_ERROR           = "failed to start HTTP server"
	SERVER_FORCED_SHUTDOWN = "server shutdown forced"
	SERVER_EXIT            = "server shutdown complete"
)

const (
	CONTROLLER_CREATE_VOLUME          = "create volume called"
	CONTROLLER_CREATE_VOLUME_FAILED   = "unable to create volume"
	CONTROLLER_GET_VOLUME             = "get volume called"
	CONTROLLER_GET_VOLUME_OK          = "volume was found"
	CONTROLLER_GET_VOLUME_FAILED      = "unable to get volume"
	CONTROLLER_VOLUME_EXISTS          = "volume exists with different size"
	CONTROLLER_VOLUME_ALREADY_CREATED = "volume already created"
	CONTROLLER_VOLUME_CREATED         = "volume was created"
	CONTROLLER_STORAGE_FULL           = "storage space full"

	CONTROLLER_LIST_VMS        = "list VMs called"
	CONTROLLER_LIST_VMS_FAILED = "list VMs failed"
	CONTROLLER_VMS_LISTED      = "VMs were listed"

	CONTROLLER_GET_VM        = "get VM called"
	CONTROLLER_GET_VM_FAILED = "get VM failed"
	CONTROLLER_GOT_VM        = "got VM"

	CONTROLLER_VOLUME_DELETE        = "delete volume called"
	CONTROLLER_VOLUME_DELETED       = "volume was deleted"
	CONTROLLER_VOLUME_DELETE_FAILED = "unable to delete volume"

	CONTROLLER_LIST_VOLUMES        = "list volumes called"
	CONTROLLER_LIST_VOLUMES_FAILED = "cannot list volumes"
	CONTROLLER_VOLUMES_LISTED      = "volumes were listed"

	CONTROLLER_GET_CAPACITY        = "get capacity called"
	CONTROLLER_GET_CAPACITY_FAILED = "cannot get capacity"
	CONTROLLER_GOT_CAPACITY        = "got storage capacity"

	CONTROLLER_PUBLISH_VOLUME        = "publish volume called"
	CONTROLLER_PUBLISH_VOLUME_FAILED = "unable to publish volume"
	CONTROLLER_VOLUME_PUBLISHED      = "volume was pubished"

	CONTROLLER_UNPUBLISH_VOLUME        = "unpublish volume called"
	CONTROLLER_UNPUBLISH_VOLUME_FAILED = "unable to unmount volume"
	CONTROLLER_VOLUME_UNPUBLISHED      = "volume was unmounted"
)
