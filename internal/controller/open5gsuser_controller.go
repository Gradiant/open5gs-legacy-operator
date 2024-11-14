package controller

import (
	"context"
	"fmt"
	netv1 "gradiant/open5gs-legacy-operator/api/v1"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Core resources
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// net group resources
//+kubebuilder:rbac:groups=net.gradiant.org,resources=open5gs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=net.gradiant.org,resources=open5gs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=net.gradiant.org,resources=open5gs/finalizers,verbs=update
//+kubebuilder:rbac:groups=net.gradiant.org,resources=open5gsusers,verbs=get;list;watch;create;update;patch;delete

// Apps resources
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// Batch resources
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// RBAC resources
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Autoscaling resources
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

// Policy resources
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

// Network resources
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

// Monitoring resources
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

// Helm-specific resources
//+kubebuilder:rbac:groups=helm.toolkit.fluxcd.io,resources=helmreleases,verbs=get;list;watch;create;update;patch;delete

const (
	Open5GSUserFinalizer = "finalizer.open5gsuser.net.gradiant.org/user"
	lastAppliedConfig    = "kubectl.kubernetes.io/last-applied-configuration"
)

type Open5GSUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Open5GSUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var Open5GSUser netv1.Open5GSUser
	var ipService string
	var err error
	if err := r.Get(ctx, req.NamespacedName, &Open5GSUser); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Printf("Error retrieving Open5GSUser %s: %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	}
	serviceName := fmt.Sprintf("%s-mongodb", strings.ToLower(Open5GSUser.Spec.Open5GS.Name))
	if Open5GSUser.Spec.Open5GS.Namespace == "" {
		ipService, err = r.GetServiceIp(ctx, serviceName, Open5GSUser.Namespace)
	} else {
		ipService, err = r.GetServiceIp(ctx, serviceName, Open5GSUser.Spec.Open5GS.Namespace)
	}

	if err != nil {
		log.Printf("Failed to get IP for service %s: %v", serviceName, err)
		return ctrl.Result{}, err
	}

	mongoURI := "mongodb://" + ipService + ":27017"

	if Open5GSUser.ObjectMeta.DeletionTimestamp.IsZero() {
		// Resource is not being deleted
		if !containsString(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer) {
			log.Printf("Creating subscriber IMSI: %s", Open5GSUser.Spec.IMSI)
			if Open5GSUser.Spec.SST != "" || Open5GSUser.Spec.SD != "" {
				// SST or SD specified, call function handling slice.
				if err := addSubscriberWithSlice(Open5GSUser, mongoURI); err != nil {
					log.Printf("Failed to create subscriber with slice %s", err)
					return ctrl.Result{}, err
				}
				log.Printf("Subscriber with slice created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
				Open5GSUser.ObjectMeta.Finalizers = append(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer)
				if err := r.Update(ctx, &Open5GSUser); err != nil {
					return ctrl.Result{}, err
				}
			} else if Open5GSUser.Spec.APN != "" {
				// APN specified without SST and SD, use function for default APN.
				if err := addSubscriberWithAPN(Open5GSUser, mongoURI); err != nil {
					log.Printf("Failed to create subscriber with APN %s", err)
					return ctrl.Result{}, err
				}
				log.Printf("Subscriber with APN created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
				Open5GSUser.ObjectMeta.Finalizers = append(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer)
				if err := r.Update(ctx, &Open5GSUser); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				// If SST, SD, and APN not specified but Key and OPC are, use defaults.
				if Open5GSUser.Spec.Key != "" && Open5GSUser.Spec.OPC != "" {
					if err := addSubscriberWithDefaults(Open5GSUser, mongoURI); err != nil {
						log.Printf("Failed to create subscriber with defaults %s", err)
						return ctrl.Result{}, err
					}
					log.Printf("Subscriber with defaults created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
					Open5GSUser.ObjectMeta.Finalizers = append(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer)
					if err := r.Update(ctx, &Open5GSUser); err != nil {
						return ctrl.Result{}, err
					}
				} else {
					// Handle the case where necessary data is not provided.
					log.Printf("Cannot create subscriber, missing key data IMSI: %s", Open5GSUser.Spec.IMSI)
					return ctrl.Result{}, fmt.Errorf("insufficient data provided for subscriber creation")
				}
			}
		} else {
			// Resource already has the finalizer
			var subscriber bson.M
			clientOptions := options.Client().ApplyURI(mongoURI)
			client, err := mongo.Connect(ctx, clientOptions)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to connect to mongo: %v", err)
			}
			defer client.Disconnect(ctx)
			collection := client.Database("open5gs").Collection("subscribers")

			err = collection.FindOne(ctx, bson.M{"imsi": Open5GSUser.Spec.IMSI}).Decode(&subscriber)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					log.Printf("Subscriber %s not found in database, creating it", Open5GSUser.Spec.IMSI)
					if Open5GSUser.Spec.SST != "" || Open5GSUser.Spec.SD != "" {
						// SST or SD specified, call function handling slice.
						if err := addSubscriberWithSlice(Open5GSUser, mongoURI); err != nil {
							log.Printf("Failed to create subscriber with slice %s", err)
							return ctrl.Result{}, err
						}
						log.Printf("Subscriber with slice created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
					} else if Open5GSUser.Spec.APN != "" {
						// APN specified without SST and SD, use function for default APN.
						if err := addSubscriberWithAPN(Open5GSUser, mongoURI); err != nil {
							log.Printf("Failed to create subscriber with APN %s", err)
							return ctrl.Result{}, err
						}
						log.Printf("Subscriber with APN created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
					} else {
						// If SST, SD, and APN not specified but Key and OPC are, use defaults.
						if Open5GSUser.Spec.Key != "" && Open5GSUser.Spec.OPC != "" {
							if err := addSubscriberWithDefaults(Open5GSUser, mongoURI); err != nil {
								log.Printf("Failed to create subscriber with defaults %s", err)
								return ctrl.Result{}, err
							}
							log.Printf("Subscriber with defaults created successfully IMSI: %s", Open5GSUser.Spec.IMSI)
						} else {
							// Handle the case where necessary data is not provided.
							log.Printf("Cannot create subscriber, missing key data IMSI: %s", Open5GSUser.Spec.IMSI)
							return ctrl.Result{}, fmt.Errorf("insufficient data provided for subscriber creation")
						}
					}
				} else {
					log.Printf("Failed to find subscriber %s in database: %v", Open5GSUser.Spec.IMSI, err)
					return ctrl.Result{}, err
				}
			} else {
				if hasDrift(Open5GSUser, subscriber) {
					log.Printf("Change detected for subscriber %s, updating it", Open5GSUser.Spec.IMSI)
					if err := updateSubscriber(Open5GSUser, mongoURI); err != nil {
						log.Printf("Failed to correct drift for subscriber %s: %v", Open5GSUser.Spec.IMSI, err)
						return ctrl.Result{}, err
					}
				}
			}
		}
	} else {
		if containsString(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer) {
			log.Printf("Deleting subscriber for Open5GSUser %s", Open5GSUser.Name)
			if err := deleteSubscriber(Open5GSUser, mongoURI); err != nil {
				log.Printf("Failed to delete subscriber for Open5GSUser %s: %v", Open5GSUser.Name, err)
				return ctrl.Result{}, err
			}
			Open5GSUser.ObjectMeta.Finalizers = removeString(Open5GSUser.ObjectMeta.Finalizers, Open5GSUserFinalizer)
			if err := r.Update(ctx, &Open5GSUser); err != nil {
				log.Printf("Failed to remove finalizer for Open5GSUser %s: %v", Open5GSUser.Name, err)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil // Requeue every 15 seconds
}

func (r *Open5GSUserReconciler) GetServiceIp(ctx context.Context, serviceName string, namespace string) (string, error) {
	var service corev1.Service
	namespacedName := client.ObjectKey{Name: serviceName, Namespace: namespace}
	if err := r.Get(ctx, namespacedName, &service); err != nil {
		log.Printf("Failed to get service %s: %v", serviceName, err)
		return "", err
	}
	return service.Spec.ClusterIP, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return
}

func (r *Open5GSUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Open5GSUser{}).
		Complete(r)
}
