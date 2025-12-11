## Related Repositories

<table>
<thead>
  <tr>
    <th align="center" style="text-align:center;width=100;">플랫폼</th>
    <th align="center" colspan="2" style="text-align:center; width=100;">컨테이너 플랫폼</th>
    <th align="center" style="text-align:center;width=250;">사이드카</th>
  </tr></thead>
<tbody>
  <tr>
    <td align="center">포털</td>
    <td align="center" colspan="2"><a href="https://github.com/K-PaaS/cp-portal-release">CP 포털</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/sidecar-deployment/tree/master/install-scripts/portal">사이드카 포털</a></td>
  </tr>
  <tr>
    <td rowspan="8">Component<br>/서비스</td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-portal-ui">Portal UI</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-remote-api">Remote API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/sidecar-portal-ui">Portal UI</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-portal-api">Portal API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-migration-ui">Migration UI</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/sidecar-portal-api">Portal API</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-portal-common-api">Common API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-migration-api">Migration API</a></td>
    <td align="center"></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-metrics-api">Metric API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-migration-auth-api">Migration Auth API</a></td>
    <td align="center"></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-terraman">Terraman API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-federation-ui">Federation UI</a></td>
    <td align="center"></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-catalog-api">Catalog API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-federation-api">🚩Federation API</a></td>
    <td align="center"></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/K-PaaS/cp-chaos-api">Chaos API</a></td>
    <td align="center"><a href="https://github.com/K-PaaS/cp-federation-collector">Federation Collector</a></td>
    <td align="center"></td>
  </tr>
  <tr>
  <td align="center">
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://github.com/K-PaaS/cp-chaos-collector">Chaos Collector</a>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  </td>
  <td align="center"></td>
  <td align="center"></td>
  </tr>
</tbody></table>
<i>🚩 You are here.</i>
<br>
<br>

## K-PaaS 컨테이너 플랫폼 Federation API
페더레이션 기반으로 호스트·멤버 클러스터 전반에 걸친 리소스 전파 및 관리 기능을 제공합니다.
- [시작하기](#시작하기)
  - [cp-federation-api 빌드 방법](#cp-federation-api-빌드-방법)
- [문서](#문서)
- [개발 환경](#개발-환경)
- [라이선스](#라이선스)

<br> 

## 시작하기
K-PaaS 컨테이너 플랫폼 Federation API가 수행하는 관리 작업은 다음과 같습니다.
- 호스트·멤버 클러스터 관리
- PropagationPolicy 관리
- ClusterPropagationPolicy 관리
- 전파 Resource 관리
- Sync 관리 

<br>

#### cp-federation-api 빌드 방법
cp-federation-api 소스 코드를 활용하여 로컬 환경에서 빌드가 필요한 경우 다음 명령어를 입력합니다.
```
$ go build
```

<br>

## 문서
- 컨테이너 플랫폼 활용에 대한 정보는 [K-PaaS 컨테이너 플랫폼](https://github.com/K-PaaS/container-platform)을 참조하십시오.

<br>

## 개발 환경
K-PaaS 컨테이너 플랫폼 Federation API의 개발 환경은 다음과 같습니다.

| Dependencies                | Version |
|-----------------------------| ------- |
| go                          | 1.24    |
| gin-contrib/cors            | v1.7.5  |
| gin-gonic/gin               | v1.10.0 |
| golang-jwt/jwt/v5           | v5.2.2  |
| karmada-io/karmada          | v1.13.1 |
| nats-io/nats                | v1.43.0 |
| spf13/viper                 | v1.18.2 |
| k8s.io/api                  | v0.32.3 |
| k8s.io/apimachinery         | v0.32.3 |

<br>

## 라이선스
Federation API는 [Apache-2.0 License](http://www.apache.org/licenses/LICENSE-2.0)를 사용합니다.