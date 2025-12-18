package service

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/api/market/resource"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"

	"github.com/gin-gonic/gin"
)

// ResourceService provides resource management functionality
type ResourceService struct {
}

// NewResourceService creates a new ResourceService instance
func NewResourceService() *ResourceService {
	return &ResourceService{}
}

// convertPVCInfo converts PVC information to protobuf format
func convertPVCInfo(k8sPVC *k8s.PVCInfo) *resource.PVCInfo {
	return &resource.PVCInfo{
		Name:         k8sPVC.Name,
		Namespace:    k8sPVC.Namespace,
		Status:       k8sPVC.Status,
		VolumeName:   k8sPVC.VolumeName,
		StorageClass: k8sPVC.StorageClass,
		Capacity:     k8sPVC.Capacity,
		AccessModes:  k8sPVC.AccessModes,
		Labels:       k8sPVC.Labels,
		CreationTime: k8sPVC.CreationTime,
		Pods:         k8sPVC.Pods,
	}
}

// convertNodeInfo converts node information to protobuf format
func convertNodeInfo(k8sNode k8s.NodeInfo) *resource.NodeInfo {
	return &resource.NodeInfo{
		Name:              k8sNode.Name,
		Status:            k8sNode.Status,
		Roles:             k8sNode.Roles,
		Version:           k8sNode.Version,
		InternalIp:        k8sNode.InternalIP,
		ExternalIp:        k8sNode.ExternalIP,
		OperatingSystem:   k8sNode.OperatingSystem,
		Architecture:      k8sNode.Architecture,
		KernelVersion:     k8sNode.KernelVersion,
		ContainerRuntime:  k8sNode.ContainerRuntime,
		AllocatableMemory: k8sNode.AllocatableMemory,
		AllocatableCpu:    k8sNode.AllocatableCPU,
		AllocatablePods:   k8sNode.AllocatablePods,
		Labels:            k8sNode.Labels,
		Annotations:       k8sNode.Annotations,
		CreationTime:      k8sNode.CreationTime,
	}
}

// convertStorageClassInfo converts storage class information to protobuf format
func convertStorageClassInfo(k8sSC k8s.StorageClassInfo) *resource.StorageClassInfo {
	return &resource.StorageClassInfo{
		Name:                 k8sSC.Name,
		Provisioner:          k8sSC.Provisioner,
		ReclaimPolicy:        k8sSC.ReclaimPolicy,
		VolumeBindingMode:    k8sSC.VolumeBindingMode,
		Parameters:           k8sSC.Parameters,
		AllowVolumeExpansion: k8sSC.AllowVolumeExpansion,
		MountOptions:         k8sSC.MountOptions,
	}
}

// ListPVCsHandler handles PVC listing requests
func (s *ResourceService) ListPVCsHandler(c *gin.Context) {
	// Get environment ID parameter
	var req resource.ListPVCsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// Use ResourceService to handle request
	result, err := s.ListPVCs(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListPVCs retrieves a list of PVCs
func (s *ResourceService) ListPVCs(req *resource.ListPVCsRequest) (*resource.ListPVCsResponse, error) {
	// Use data processing layer to get PVC list
	pvcList, err := biz.GResourceBiz.ListPVCs(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get PVC list: %s", err.Error())
	}

	// Convert to protobuf type
	var pbPVCList []*resource.PVCInfo
	for _, pvc := range pvcList {
		pbPVCList = append(pbPVCList, convertPVCInfo(&pvc))
	}

	// Build response
	response := &resource.ListPVCsResponse{
		List: pbPVCList,
	}

	return response, nil
}

// convertDockerVolumeInfo converts Docker volume info to protobuf format
func convertDockerVolumeInfo(v container.VolumeInfo) *resource.DockerVolumeInfo {
	statusMap := make(map[string]string)
	for k, val := range v.Status {
		statusMap[k] = fmt.Sprintf("%v", val)
	}

	return &resource.DockerVolumeInfo{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		Labels:     v.Labels,
		Options:    v.Options,
		Scope:      v.Scope,
		CreatedAt:  v.CreatedAt,
		Status:     statusMap,
	}
}

// ListDockerVolumesHandler Docker volume list handler
func (s *ResourceService) ListDockerVolumesHandler(c *gin.Context) {
	var req resource.ListDockerVolumesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	result, err := s.ListDockerVolumes(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListDockerVolumes lists Docker volumes
func (s *ResourceService) ListDockerVolumes(req *resource.ListDockerVolumesRequest) (*resource.ListDockerVolumesResponse, error) {
	list, err := biz.GResourceBiz.ListDockerVolumes(uint(req.EnvironmentId))
	if err != nil {
		return nil, err
	}

	var pbList []*resource.DockerVolumeInfo
	for _, v := range list {
		pbList = append(pbList, convertDockerVolumeInfo(v))
	}

	return &resource.ListDockerVolumesResponse{
		List: pbList,
	}, nil
}

// CreateDockerVolumeHandler create Docker volume handler
func (s *ResourceService) CreateDockerVolumeHandler(c *gin.Context) {
	var req resource.CreateDockerVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "request parameter error: "+err.Error())
		return
	}

	if config.IsDemoMode() {
		common.GinError(c, i18nresp.CodeForbidden, "operation forbidden in demo mode")
		return
	}

	result, err := s.CreateDockerVolume(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// CreateDockerVolume creates a Docker volume
func (s *ResourceService) CreateDockerVolume(req *resource.CreateDockerVolumeRequest) (*resource.CreateDockerVolumeResponse, error) {
	vol, err := biz.GResourceBiz.CreateDockerVolume(uint(req.EnvironmentId), req.Name, req.Labels)
	if err != nil {
		return nil, err
	}

	return &resource.CreateDockerVolumeResponse{
		Volume: convertDockerVolumeInfo(vol),
	}, nil
}

// FindDockerVolumeHandler get Docker volume handler
func (s *ResourceService) FindDockerVolumeHandler(c *gin.Context) {
	var req resource.FindDockerVolumeRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	result, err := s.FindDockerVolume(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// FindDockerVolume inspects a Docker volume
func (s *ResourceService) FindDockerVolume(req *resource.FindDockerVolumeRequest) (*resource.FindDockerVolumeResponse, error) {
	vol, err := biz.GResourceBiz.GetDockerVolume(uint(req.EnvironmentId), req.Name)
	if err != nil {
		return nil, err
	}

	return &resource.FindDockerVolumeResponse{
		Volume: convertDockerVolumeInfo(vol),
	}, nil
}

// RemoveDockerVolumeHandler remove Docker volume handler
func (s *ResourceService) RemoveDockerVolumeHandler(c *gin.Context) {
	var req resource.RemoveDockerVolumeRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	if config.IsDemoMode() {
		common.GinError(c, i18nresp.CodeForbidden, "operation forbidden in demo mode")
		return
	}

	result, err := s.RemoveDockerVolume(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// RemoveDockerVolume removes a Docker volume
func (s *ResourceService) RemoveDockerVolume(req *resource.RemoveDockerVolumeRequest) (*resource.RemoveDockerVolumeResponse, error) {
	err := biz.GResourceBiz.RemoveDockerVolume(uint(req.EnvironmentId), req.Name)
	if err != nil {
		return nil, err
	}

	return &resource.RemoveDockerVolumeResponse{
		Success: true,
	}, nil
}

// CreatePVCHandler create PVC interface Handler
func (s *ResourceService) CreatePVCHandler(c *gin.Context) {
	var req resource.CreatePVCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "request parameter error: "+err.Error())
		return
	}

	// Block PVC creation in demo mode
	if config.IsDemoMode() {
		common.GinError(c, i18nresp.CodeForbidden, "operation forbidden in demo mode")
		return
	}

	// Use ResourceService to handle request
	result, err := s.CreatePVC(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// CreatePVC creates a new PVC
func (s *ResourceService) CreatePVC(req *resource.CreatePVCRequest) (*resource.CreatePVCResponse, error) {
	// Validate required parameters
	if req.Name == "" {
		return nil, fmt.Errorf("PVC name cannot be empty")
	}
	if req.EnvironmentId <= 0 {
		return nil, fmt.Errorf("environment ID must be greater than 0")
	}
	if req.StorageSize <= 0 {
		return nil, fmt.Errorf("storage size must be greater than 0")
	}

	// Validate access mode
	validAccessModes := map[string]bool{
		"ReadWriteOnce": true,
		"ReadOnlyMany":  true,
		"ReadWriteMany": true,
	}
	if req.AccessMode != "" && !validAccessModes[req.AccessMode] {
		return nil, fmt.Errorf("invalid access mode, supported: ReadWriteOnce, ReadOnlyMany, ReadWriteMany")
	}

	var pvcInfo *k8s.PVCInfo
	var err error

	// Choose different creation methods based on whether hostPath is provided
	if req.HostPath != "" {
		// Create PVC based on host path
		if req.NodeName == "" {
			return nil, fmt.Errorf("node name cannot be empty when creating HostPath type PVC")
		}
		pvcInfo, err = biz.GResourceBiz.CreateHostPathPVC(
			uint(req.EnvironmentId),
			req.Name,
			req.HostPath,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
		)
	} else {
		// Create regular PVC
		pvcInfo, err = biz.GResourceBiz.CreatePVC(
			uint(req.EnvironmentId),
			req.Name,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
			nil, // labels
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %s", err.Error())
	}

	// Convert to protobuf type and return
	return &resource.CreatePVCResponse{
		Pvc: convertPVCInfo(pvcInfo),
	}, nil
}

// ListNodesHandler get node list interface Handler
func (s *ResourceService) ListNodesHandler(c *gin.Context) {
	// Get environment ID parameter
	var req resource.ListNodesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get node list: %s", err.Error()))
		return
	}

	// Use ResourceService to handle request
	result, err := s.ListNodes(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListNodes get node list business logic
func (s *ResourceService) ListNodes(req *resource.ListNodesRequest) (*resource.ListNodesResponse, error) {
	// Use data processing layer to get node list
	nodeList, err := biz.GResourceBiz.ListNodes(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get node list: %s", err.Error())
	}

	// Convert to protobuf type
	var pbNodeList []*resource.NodeInfo
	for _, node := range nodeList {
		pbNodeList = append(pbNodeList, convertNodeInfo(node))
	}

	// Build response
	response := &resource.ListNodesResponse{
		List: pbNodeList,
	}

	return response, nil
}

// ListStorageClassesHandler get storage class list interface Handler
func (s *ResourceService) ListStorageClassesHandler(c *gin.Context) {
	// Get environment ID parameter
	var req resource.ListStorageClassesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// Use ResourceService to handle request
	result, err := s.ListStorageClasses(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListStorageClasses get storage class list business logic
func (s *ResourceService) ListStorageClasses(req *resource.ListStorageClassesRequest) (*resource.ListStorageClassesResponse, error) {
	// Use data processing layer to get storage class list
	storageClassList, err := biz.GResourceBiz.ListStorageClasses(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get storage class list: %s", err.Error())
	}

	// Convert to protobuf type
	var pbStorageClassList []*resource.StorageClassInfo
	for _, sc := range storageClassList {
		pbStorageClassList = append(pbStorageClassList, convertStorageClassInfo(sc))
	}

	// Build response
	response := &resource.ListStorageClassesResponse{
		List: pbStorageClassList,
	}

	return response, nil
}
