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
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
)

// EnsureMemberClusterMiddleware ensures that the member cluster exists.
func EnsureMemberClusterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		karmadaClient := client.InClusterKarmadaClient()
		_, err := karmadaClient.ClusterV1alpha1().Clusters().Get(context.TODO(), c.Param("clustername"), metav1.GetOptions{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.BaseResponse{
				Code: 500,
				Msg:  err.Error(),
			})
			return
		}
		c.Next()
	}
}

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		langHeader := c.GetHeader("Accept-Language")
		lang := "en"

		if langHeader != "" {
			langParts := strings.Split(langHeader, ",")
			lang = strings.ToLower(strings.TrimSpace(langParts[0]))

			supportedLangs := map[string]bool{
				"en": true,
				"ko": true,
			}

			if _, ok := supportedLangs[lang]; !ok {
				lang = "en"
			}
		}

		c.Set("lang", lang)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	})
}
