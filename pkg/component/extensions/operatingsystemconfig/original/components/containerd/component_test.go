// Copyright 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package containerd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components"
	. "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/containerd"
	"github.com/gardener/gardener/pkg/utils"
)

var _ = Describe("Component", func() {
	Describe("#Config", func() {
		var component components.Component

		BeforeEach(func() {
			component = New()
		})

		It("should return the expected units and files", func() {
			units, files, err := component.Config(components.Context{})

			Expect(err).NotTo(HaveOccurred())

			monitorUnit := extensionsv1alpha1.Unit{
				Name:    "containerd-monitor.service",
				Command: ptr.To(extensionsv1alpha1.CommandStart),
				Enable:  ptr.To(true),
				Content: ptr.To(`[Unit]
Description=Containerd-monitor daemon
After=containerd.service
[Install]
WantedBy=multi-user.target
[Service]
Restart=always
EnvironmentFile=/etc/environment
ExecStart=/opt/bin/health-monitor-containerd`),
				FilePaths: []string{"/opt/bin/health-monitor-containerd"},
			}

			logrotateUnit := extensionsv1alpha1.Unit{
				Name:   "containerd-logrotate.service",
				Enable: ptr.To(true),
				Content: ptr.To(`[Unit]
Description=Rotate and Compress System Logs
[Service]
ExecStart=/usr/sbin/logrotate -s /var/lib/containerd-logrotate.status /etc/systemd/containerd.conf
[Install]
WantedBy=multi-user.target`),
				FilePaths: []string{"/etc/systemd/containerd.conf"},
			}

			logrotateTimerUnit := extensionsv1alpha1.Unit{
				Name:    "containerd-logrotate.timer",
				Command: ptr.To(extensionsv1alpha1.CommandStart),
				Enable:  ptr.To(true),
				Content: ptr.To(`[Unit]
Description=Log Rotation at each 10 minutes
[Timer]
OnCalendar=*:0/10
AccuracySec=1min
Persistent=true
[Install]
WantedBy=multi-user.target`),
			}

			monitorFile := extensionsv1alpha1.File{
				Path:        "/opt/bin/health-monitor-containerd",
				Permissions: ptr.To[int32](0755),
				Content: extensionsv1alpha1.FileContent{
					Inline: &extensionsv1alpha1.FileContentInline{
						Encoding: "b64",
						Data:     utils.EncodeBase64([]byte(healthMonitorScript)),
					},
				},
			}

			logrotateConfigFile := extensionsv1alpha1.File{
				Path:        "/etc/systemd/containerd.conf",
				Permissions: ptr.To[int32](0644),
				Content: extensionsv1alpha1.FileContent{
					Inline: &extensionsv1alpha1.FileContentInline{
						Data: logRotateData,
					},
				},
			}

			Expect(units).To(ConsistOf(monitorUnit, logrotateUnit, logrotateTimerUnit))
			Expect(files).To(ConsistOf(monitorFile, logrotateConfigFile))
		})
	})
})

const (
	healthMonitorScript = `#!/bin/bash
set -o nounset
set -o pipefail

function containerd_monitoring {
  echo "containerd monitor has started !"
  while [ 1 ]; do
    start_timestamp=$(date +%s)
    until ctr c list > /dev/null; do
      CONTAINERD_PID=$(systemctl show --property MainPID containerd --value)

      if [ $CONTAINERD_PID -eq 0 ]; then
          echo "Connection to containerd socket failed (process not started), retrying in $SLEEP_SECONDS seconds..."
          break
      fi

      now=$(date +%s)
      time_elapsed="$(($now-$start_timestamp))"

      if [ $time_elapsed -gt 60 ]; then
        echo "containerd daemon unreachable for more than 60s. Sending SIGTERM to PID $CONTAINERD_PID"
        kill -n 15 $CONTAINERD_PID
        sleep 20
        break 2
      fi
      echo "Connection to containerd socket failed, retrying in $SLEEP_SECONDS seconds..."
      sleep $SLEEP_SECONDS
    done
    sleep $SLEEP_SECONDS
  done
}

SLEEP_SECONDS=10
echo "Start health monitoring for containerd"
containerd_monitoring
`

	logRotateData = `/var/log/pods/*/*/*.log {
    rotate 14
    copytruncate
    missingok
    notifempty
    compress
    daily
    dateext
    dateformat -%Y%m%d-%s
    create 0644 root root
}
`
)
