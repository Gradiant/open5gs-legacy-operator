package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	netv1 "gradiant/open5gs-legacy-operator/api/v1"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func updateSubscriber(Open5GSUser netv1.Open5GSUser, mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")

	updateFields := bson.M{
		"security.k":   Open5GSUser.Spec.Key,
		"security.opc": Open5GSUser.Spec.OPC,
	}

	if Open5GSUser.Spec.SST != "" || Open5GSUser.Spec.SD != "" {
		sst, err := strconv.Atoi(Open5GSUser.Spec.SST)
		if err != nil {
			return fmt.Errorf("failed to convert SST to int: %v", err)
		}
		updateFields["slice.0.sst"] = sst
		updateFields["slice.0.sd"] = Open5GSUser.Spec.SD
	}

	updateFields["slice.0.session.0.name"] = Open5GSUser.Spec.APN

	update := bson.M{"$set": updateFields}
	filter := bson.M{"imsi": Open5GSUser.Spec.IMSI}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update subscriber: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no subscriber found with IMSI %s", Open5GSUser.Spec.IMSI)
	}

	return nil
}

func deleteSubscriber(Open5GSUser netv1.Open5GSUser, mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")

	result, err := collection.DeleteOne(ctx, bson.M{"imsi": Open5GSUser.Spec.IMSI})
	if err != nil {
		return fmt.Errorf("failed to delete subscriber: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no subscriber found with IMSI %s", Open5GSUser.Spec.IMSI)
	}

	return nil
}

func addSubscriberWithSlice(Open5GSUser netv1.Open5GSUser, mongoURI string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")

	sst, err := strconv.Atoi(Open5GSUser.Spec.SST)
	if err != nil {
		return fmt.Errorf("failed to convert SST to int: %v", err)
	}

	subscriber := bson.M{
		"_id":            primitive.NewObjectID(),
		"schema_version": 1,
		"imsi":           Open5GSUser.Spec.IMSI,
		"msisdn":         []string{},
		"imeisv":         []string{},
		"mme_host":       []string{},
		"mm_realm":       []string{},
		"purge_flag":     []string{},
		"slice": []bson.M{
			{
				"sst":               sst,
				"sd":                Open5GSUser.Spec.SD,
				"default_indicator": true,
				"session": []bson.M{
					{
						"name": Open5GSUser.Spec.APN,
						"type": 3,
						"qos": bson.M{
							"index": 9,
							"arp": bson.M{
								"priority_level":            8,
								"pre_emption_capability":    1,
								"pre_emption_vulnerability": 2,
							},
						},
						"ambr": bson.M{
							"downlink": bson.M{"value": 1000000000, "unit": 0},
							"uplink":   bson.M{"value": 1000000000, "unit": 0},
						},
						"pcc_rule": []string{},
						"_id":      primitive.NewObjectID(),
					},
				},
				"_id": primitive.NewObjectID(),
			},
		},
		"security": bson.M{
			"k":   Open5GSUser.Spec.Key,
			"opc": Open5GSUser.Spec.OPC,
			"amf": "8000",
		},
		"ambr": bson.M{
			"downlink": bson.M{"value": 1000000000, "unit": 0},
			"uplink":   bson.M{"value": 1000000000, "unit": 0},
		},
		"access_restriction_data":     32,
		"network_access_mode":         0,
		"subscriber_status":           0,
		"operator_determined_barring": 0,
		"subscribed_rau_tau_timer":    12,
		"__v":                         0,
	}

	_, err = collection.InsertOne(ctx, subscriber)
	if err != nil {
		return fmt.Errorf("failed to insert subscriber: %v", err)
	}

	return nil
}

func addSubscriberWithAPN(Open5GSUser netv1.Open5GSUser, mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")

	const defaultSST = 1

	subscriber := bson.M{
		"_id":            primitive.NewObjectID(),
		"schema_version": 1,
		"imsi":           Open5GSUser.Spec.IMSI,
		"msisdn":         []string{},
		"imeisv":         []string{},
		"mme_host":       []string{},
		"mm_realm":       []string{},
		"purge_flag":     []string{},
		"slice": []bson.M{
			{
				"sst":               defaultSST,
				"default_indicator": true,
				"session": []bson.M{
					{
						"name": Open5GSUser.Spec.APN,
						"type": 3,
						"qos": bson.M{
							"index": 9,
							"arp": bson.M{
								"priority_level":            8,
								"pre_emption_capability":    1,
								"pre_emption_vulnerability": 2,
							},
						},
						"ambr": bson.M{
							"downlink": bson.M{"value": 1000000000, "unit": 0},
							"uplink":   bson.M{"value": 1000000000, "unit": 0},
						},
						"pcc_rule": []string{},
						"_id":      primitive.NewObjectID(),
					},
				},
				"_id": primitive.NewObjectID(),
			},
		},
		"security": bson.M{
			"k":   Open5GSUser.Spec.Key,
			"opc": Open5GSUser.Spec.OPC,
			"amf": "8000",
		},
		"ambr": bson.M{
			"downlink": bson.M{"value": 1000000000, "unit": 0},
			"uplink":   bson.M{"value": 1000000000, "unit": 0},
		},
		"access_restriction_data":     32,
		"network_access_mode":         0,
		"subscriber_status":           0,
		"operator_determined_barring": 0,
		"subscribed_rau_tau_timer":    12,
		"__v":                         0,
	}

	_, err = collection.InsertOne(ctx, subscriber)
	if err != nil {
		return fmt.Errorf("failed to insert subscriber: %v", err)
	}

	return nil
}
func addSubscriberWithDefaults(Open5GSUser netv1.Open5GSUser, mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")

	const defaultSST = 1
	const defaultAPN = "internet"

	subscriber := bson.M{
		"_id":            primitive.NewObjectID(),
		"schema_version": 1,
		"imsi":           Open5GSUser.Spec.IMSI,
		"msisdn":         []string{},
		"imeisv":         []string{},
		"mme_host":       []string{},
		"mm_realm":       []string{},
		"purge_flag":     []string{},
		"slice": []bson.M{
			{
				"sst":               defaultSST,
				"default_indicator": true,
				"session": []bson.M{
					{
						"name": defaultAPN,
						"type": 3,
						"qos": bson.M{
							"index": 9,
							"arp": bson.M{
								"priority_level":            8,
								"pre_emption_capability":    1,
								"pre_emption_vulnerability": 2,
							},
						},
						"ambr": bson.M{
							"downlink": bson.M{"value": 1000000000, "unit": 0},
							"uplink":   bson.M{"value": 1000000000, "unit": 0},
						},
						"pcc_rule": []string{},
						"_id":      primitive.NewObjectID(),
					},
				},
				"_id": primitive.NewObjectID(),
			},
		},
		"security": bson.M{
			"k":   Open5GSUser.Spec.Key,
			"opc": Open5GSUser.Spec.OPC,
			"amf": "8000",
		},
		"ambr": bson.M{
			"downlink": bson.M{"value": 1000000000, "unit": 0},
			"uplink":   bson.M{"value": 1000000000, "unit": 0},
		},
		"access_restriction_data":     32,
		"network_access_mode":         0,
		"subscriber_status":           0,
		"operator_determined_barring": 0,
		"subscribed_rau_tau_timer":    12,
		"__v":                         0,
	}

	_, err = collection.InsertOne(ctx, subscriber)
	if err != nil {
		return fmt.Errorf("failed to insert subscriber with defaults: %v", err)
	}

	return nil
}

func hasDrift(open5GSUser netv1.Open5GSUser, subscriber bson.M) bool {
	if subscriber["security"].(bson.M)["k"] != open5GSUser.Spec.Key {
		return true
	}
	if subscriber["security"].(bson.M)["opc"] != open5GSUser.Spec.OPC {
		return true
	}

	if sst, ok := subscriber["slice"].(bson.A)[0].(bson.M)["sst"].(int); ok && strconv.Itoa(sst) != open5GSUser.Spec.SST {
		return true
	}

	if sd, ok := subscriber["slice"].(bson.A)[0].(bson.M)["sd"].(string); ok && sd != open5GSUser.Spec.SD {
		return true
	}

	if name, ok := subscriber["slice"].(bson.A)[0].(bson.M)["session"].(bson.A)[0].(bson.M)["name"].(string); ok && name != open5GSUser.Spec.APN {
		return true
	}

	return false
}

func (r *Open5GSUserReconciler) ListOpen5GSUsers(ctx context.Context) ([]netv1.Open5GSUser, error) {
	var userList netv1.Open5GSUserList
	if err := r.List(ctx, &userList); err != nil {
		return nil, err
	}
	return userList.Items, nil
}
func listSubscribers(ctx context.Context, mongoURI string) ([]bson.M, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	collection := client.Database("open5gs").Collection("subscribers")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscribers []bson.M
	if err = cursor.All(ctx, &subscribers); err != nil {
		return nil, err
	}
	return subscribers, nil
}
