#!/bin/bash

# Emergency rollback script - use simple GetAll query without optimization

echo "Creating emergency rollback for GetEventStatus..."

cat > /tmp/getstatus_simple.txt << 'EOF'
func (s *InteractionService) GetEventStatus(ctx context.Context, eventID string) (map[string]interface{}, error) {
	log.Printf("[GetEventStatus] SIMPLE MODE - Fetching ALL records for event: %s", eventID)
	
	// Check cache first
	if cached, found := s.Cache.Get(eventID); found {
		log.Printf("[GetEventStatus] Cache HIT for event: %s", eventID)
		return cached, nil
	}
	log.Printf("[GetEventStatus] Cache MISS for event: %s", eventID)

	// EMERGENCY ROLLBACK: Use simple GetAll() without WHERE/OrderBy
	recordsRef := s.Repo.Client.Collection("events").Doc(eventID).Collection("records")
	
	allRecords, err := recordsRef.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("[GetEventStatus] ERROR getting ALL records: %v", err)
		return nil, err
	}
	
	log.Printf("[GetEventStatus] Successfully fetched %d total records", len(allRecords))

	// Build result
	var result = make(map[string]interface{})
	var list []map[string]interface{}

	for _, doc := range allRecords {
		var rec models.Interaction
		doc.DataTo(&rec)

		recMap := map[string]interface{}{
			"id":              doc.Ref.ID,
			"type":            rec.Type,
			"userId":          rec.UserID,
			"userDisplayName": rec.UserDisplayName,
			"userPictureUrl":  rec.UserPictureUrl,
			"timestamp":       rec.Timestamp,
			"status":          rec.Status,
			"selectedOptions": rec.SelectedOptions,
			"count":           rec.Count,
			"note":            rec.Note,
			"content":         rec.Content,
			"clapCount":       rec.ClapCount,
		}
		list = append(list, recMap)
	}

	log.Printf("[GetEventStatus] Returning %d records for event: %s", len(list), eventID)

	result["records"] = list
	
	// Cache the result for 30 seconds
	s.Cache.Set(eventID, result)
	
	return result, nil
}
EOF

echo "Emergency rollback function saved to /tmp/getstatus_simple.txt"
echo "If needed, replace GetEventStatus in interaction.go with this simple version"
