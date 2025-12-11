/*
Copyright 2024 The Karmada Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/auth"
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/federation"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/cluster"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/clusterpropagationpolicy"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/namespace"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/overview"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/propagationpolicy"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/resources"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/sync"
	"github.com/karmada-io/dashboard/pkg/environment"
	"net/http"
)

var (
	router   *gin.Engine
	v1       *gin.RouterGroup
	member   *gin.RouterGroup
	StatusUp = "UP"
)

type HealthStatus struct {
	Status string `json:"status"`
}

func init() {
	authAdapter := auth.NewInternalAuthAdapter()
	fedAdapter := federation.NewInternalAdapter(authAdapter)

	if !environment.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()
	router.Use(CORSMiddleware())

	router.GET("/livez", func(c *gin.Context) {
		c.String(200, "livez")
	})
	router.GET("/readyz", func(c *gin.Context) {
		c.String(200, "readyz")
	})

	router.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthStatus{Status: StatusUp})
	})
	router.GET("/actuator/health/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthStatus{Status: StatusUp})
	})

	router.Use(LanguageMiddleware())
	router.Use(authAdapter.AuthMiddleware())
	_ = router.SetTrustedProxies(nil)

	v1 = router.Group("/api/v1")

	// cluster
	clusterV1 := v1.Group("")
	clusterHandler := cluster.NewHandler(fedAdapter)
	clusterV1.GET("/registrable-clusters", clusterHandler.HandleListRegisterableClusters)
	clusterV1.GET("/cluster", clusterHandler.HandleListClusters)
	clusterV1.GET("/cluster/:clusterId", clusterHandler.HandleGetClusterYaml)
	clusterV1.POST("/cluster", clusterHandler.HandlePostCluster)
	clusterV1.DELETE("/cluster/:clusterId", clusterHandler.HandleDeleteCluster)

	// overview
	overviewV1 := v1.Group("/overview")
	overviewHandler := overview.NewHandler(clusterHandler)
	overviewV1.GET("", overviewHandler.HandleGetOverview)

	// propagationPolicy
	ppV1 := v1.Group("/propagationpolicy")
	ppV1.GET("", propagationpolicy.HandleGetPropagationPolicyList)
	ppV1.GET("/namespace/:namespace/:propagationPolicyName", propagationpolicy.HandleGetPropagationPolicyYaml)
	ppV1.POST("", propagationpolicy.HandlePostPropagationPolicy)
	ppV1.PUT("/namespace/:namespace/:propagationPolicyName", propagationpolicy.HandlePutPropagationPolicy)
	ppV1.DELETE("/namespace/:namespace/:propagationPolicyName", propagationpolicy.HandleDeletePropagationPolicy)

	// clusterPropagationPolicy
	ccpV1 := v1.Group("/clusterpropagationpolicy")
	ccpV1.GET("", clusterpropagationpolicy.HandleGetClusterPropagationPolicyList)
	ccpV1.GET("/:clusterPropagationPolicyName", clusterpropagationpolicy.HandleGetClusterPropagationPolicyYaml)
	ccpV1.POST("", clusterpropagationpolicy.HandlePostClusterPropagationPolicy)
	ccpV1.PUT("/:clusterPropagationPolicyName", clusterpropagationpolicy.HandlePutClusterPropagationPolicy)
	ccpV1.DELETE("/:clusterPropagationPolicyName", clusterpropagationpolicy.HandleDeleteClusterPropagationPolicy)

	// namespace
	namespaceV1 := v1.Group("/namespace")
	namespaceV1.GET("", namespace.HandleGetNamespaceList)
	namespaceV1.GET("/:namespaceName", namespace.HandleGetNamespaceYaml)
	namespaceV1.POST("", namespace.HandleCreateNamespace)
	namespaceV1.DELETE("/:namespaceName", namespace.HandleDeleteNamespace)

	// resource
	resourceV1 := v1.Group("/resource")
	resourceV1.GET("/namespaces", resources.HandleGetResourceNamespace)
	resourceV1.GET("/names", resources.HandleGetResourceNames)
	resourceV1.GET("/labels", resources.HandleGetResourceLabels)
	resourceV1.GET("/:kind", resources.HandleListResource)
	resourceV1.GET("/:kind/namespace/:namespace/name/:name", resources.HandleGetResourceYaml)
	resourceV1.POST("", resources.HandleCreateResource)
	resourceV1.PUT("", resources.HandleUpdateResource)
	resourceV1.DELETE("/:kind/namespace/:namespace/name/:name", resources.HandleDeleteResource)

	// sync
	syncV1 := v1.Group("/sync")
	syncHandler := sync.NewHandler(fedAdapter)
	syncV1.GET("/resource/:clusterId", syncHandler.HandleGetSyncResources)
	syncV1.POST("/:clusterId", syncHandler.HandlePostSync)

	member = v1.Group("/member/:clustername")
	member.Use(EnsureMemberClusterMiddleware())

}

// V1 returns the router group for /api/v1 which for resources in control plane endpoints.
func V1() *gin.RouterGroup {
	return v1
}

// Router returns the main Gin engine instance.
func Router() *gin.Engine {
	return router
}

// MemberV1 returns the router group for /api/v1/member/:clustername which for resources in specific member cluster.
func MemberV1() *gin.RouterGroup {
	return member
}
