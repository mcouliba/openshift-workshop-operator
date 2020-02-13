//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package workshop

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

func (r *ReconcileWorkshop) ApproveInstallPlan(subscriptionName string, namespace string) error {

	subscription, subscriptionErr := r.GetSubscription(subscriptionName, namespace)

	if subscriptionErr != nil {
		return subscriptionErr
	}

	if subscription.Status.InstallPlanRef == nil {
		return errors.New("InstallPlan Approval: Subscription is not ready yet")
	}

	installPlan, installPlanErr := r.GetInstallPlan(subscription.Status.InstallPlanRef.Name, namespace)

	if installPlanErr != nil {
		return installPlanErr
	}

	if !installPlan.Spec.Approved {
		installPlan.Spec.Approved = true
		if err := r.client.Update(context.TODO(), installPlan); err != nil {
			return err
		}
		logrus.Infof("%s Subscription in % project Approved", subscriptionName, namespace)
	}

	return nil
}
