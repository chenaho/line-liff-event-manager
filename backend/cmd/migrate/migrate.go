package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	_ "github.com/lib/pq"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Models
type EventConfig struct {
	AllowMultiSelect   bool      `json:"allowMultiSelect,omitempty" firestore:"allowMultiSelect,omitempty"`
	MaxVotes           int       `json:"maxVotes,omitempty" firestore:"maxVotes,omitempty"`
	Options            []string  `json:"options,omitempty" firestore:"options,omitempty"`
	MaxParticipants    int       `json:"maxParticipants,omitempty" firestore:"maxParticipants,omitempty"`
	WaitlistLimit      int       `json:"waitlistLimit,omitempty" firestore:"waitlistLimit,omitempty"`
	MaxCountPerUser    int       `json:"maxCountPerUser,omitempty" firestore:"maxCountPerUser,omitempty"`
	StartTime          time.Time `json:"startTime,omitempty" firestore:"startTime,omitempty"`
	EndTime            time.Time `json:"endTime,omitempty" firestore:"endTime,omitempty"`
	MaxCommentsPerUser int       `json:"maxCommentsPerUser,omitempty" firestore:"maxCommentsPerUser,omitempty"`
	AllowReaction      bool      `json:"allowReaction,omitempty" firestore:"allowReaction,omitempty"`
}

type Event struct {
	EventID   string      `json:"eventId" firestore:"eventId"`
	Type      string      `json:"type" firestore:"type"`
	Title     string      `json:"title" firestore:"title"`
	IsActive  bool        `json:"isActive" firestore:"isActive"`
	CreatedBy string      `json:"createdBy" firestore:"createdBy"`
	CreatedAt time.Time   `json:"createdAt" firestore:"createdAt"`
	Config    EventConfig `json:"config" firestore:"config"`
}

type User struct {
	LineUserID      string    `json:"lineUserId" firestore:"lineUserId"`
	LineDisplayName string    `json:"lineDisplayName" firestore:"lineDisplayName"`
	PictureURL      string    `json:"pictureUrl" firestore:"pictureUrl"`
	Role            string    `json:"role" firestore:"role"`
	CreatedAt       time.Time `json:"createdAt" firestore:"createdAt"`
}

type Interaction struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId" firestore:"userId"`
	UserDisplayName string    `json:"userDisplayName" firestore:"userDisplayName"`
	UserPictureUrl  string    `json:"userPictureUrl" firestore:"userPictureUrl"`
	Type            string    `json:"type" firestore:"type"`
	Timestamp       time.Time `json:"timestamp" firestore:"timestamp"`
	SelectedOptions []string  `json:"selectedOptions,omitempty" firestore:"selectedOptions,omitempty"`
	Count           int       `json:"count,omitempty" firestore:"count,omitempty"`
	Status          string    `json:"status,omitempty" firestore:"status,omitempty"`
	Note            string    `json:"note,omitempty" firestore:"note,omitempty"`
	CancelledAt     time.Time `json:"cancelledAt,omitempty" firestore:"cancelledAt,omitempty"`
	Content         string    `json:"content,omitempty" firestore:"content,omitempty"`
	ClapCount       int       `json:"clapCount,omitempty" firestore:"clapCount,omitempty"`
	Reactions       []string  `json:"reactions,omitempty" firestore:"reactions,omitempty"`
}

type InteractionPayload struct {
	SelectedOptions []string `json:"selectedOptions,omitempty"`
	Count           int      `json:"count,omitempty"`
	Note            string   `json:"note,omitempty"`
	CancelledAt     string   `json:"cancelledAt,omitempty"`
	Content         string   `json:"content,omitempty"`
	ClapCount       int      `json:"clapCount,omitempty"`
	Reactions       []string `json:"reactions,omitempty"`
}

func main() {
	log.Println("=== Firestore to PostgreSQL Migration Tool ===")

	// Check arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Initialize Firestore connection
	ctx := context.Background()
	fsClient := initFirestore(ctx)
	defer fsClient.Close()

	// dry-run only needs Firestore
	if command == "dry-run" {
		dryRun(ctx, fsClient)
		log.Println("=== Dry Run Complete ===")
		return
	}

	// Other commands need PostgreSQL
	pgDB := initPostgres()
	defer pgDB.Close()

	switch command {
	case "migrate-all":
		migrateUsers(ctx, fsClient, pgDB)
		migrateEvents(ctx, fsClient, pgDB)
		migrateInteractions(ctx, fsClient, pgDB)
	case "migrate-users":
		migrateUsers(ctx, fsClient, pgDB)
	case "migrate-events":
		migrateEvents(ctx, fsClient, pgDB)
	case "migrate-interactions":
		migrateInteractions(ctx, fsClient, pgDB)
	case "verify":
		verifyMigration(ctx, fsClient, pgDB)
	default:
		printUsage()
		os.Exit(1)
	}

	log.Println("=== Migration Complete ===")
}

func printUsage() {
	fmt.Println(`
Firestore to PostgreSQL Migration Tool

Usage: go run migrate.go <command>

Commands:
  migrate-all           Migrate users, events, and interactions
  migrate-users         Migrate only users
  migrate-events        Migrate only events  
  migrate-interactions  Migrate only interactions
  verify                Verify migration counts
  dry-run               Show what would be migrated without writing

Environment Variables:
  FIREBASE_CREDENTIALS  Path to Firebase service account JSON
  POSTGRES_HOST         PostgreSQL host (default: localhost)
  POSTGRES_PORT         PostgreSQL port (default: 5433)
  POSTGRES_USER         PostgreSQL user (default: eventmanager)
  POSTGRES_PASSWORD     PostgreSQL password (required)
  POSTGRES_DB           PostgreSQL database (default: eventmanager)
`)
}

func initFirestore(ctx context.Context) *firestore.Client {
	credsPath := os.Getenv("FIREBASE_CREDENTIALS")
	if credsPath == "" {
		log.Fatal("FIREBASE_CREDENTIALS environment variable is required")
	}

	// Read credentials file to get project ID
	credsData, err := os.ReadFile(credsPath)
	if err != nil {
		log.Fatalf("Failed to read credentials file: %v", err)
	}

	var creds struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(credsData, &creds); err != nil {
		log.Fatalf("Failed to parse credentials: %v", err)
	}

	if creds.ProjectID == "" {
		log.Fatal("project_id not found in credentials file")
	}

	log.Printf("Using Firebase project: %s", creds.ProjectID)

	opt := option.WithCredentialsFile(credsPath)
	conf := &firebase.Config{ProjectID: creds.ProjectID}
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	log.Println("✓ Connected to Firestore")
	return client
}

func initPostgres() *sql.DB {
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5433")
	user := getEnvOrDefault("POSTGRES_USER", "eventmanager")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := getEnvOrDefault("POSTGRES_DB", "eventmanager")

	if password == "" {
		log.Fatal("POSTGRES_PASSWORD environment variable is required")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	log.Println("✓ Connected to PostgreSQL")
	return db
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func migrateUsers(ctx context.Context, fs *firestore.Client, pg *sql.DB) {
	log.Println("\n--- Migrating Users ---")

	iter := fs.Collection("users").Documents(ctx)
	count := 0
	errors := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading user: %v", err)
			errors++
			continue
		}

		var user User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error parsing user %s: %v", doc.Ref.ID, err)
			errors++
			continue
		}

		// Use document ID as LineUserID if not set
		if user.LineUserID == "" {
			user.LineUserID = doc.Ref.ID
		}

		err = insertUser(pg, &user)
		if err != nil {
			log.Printf("Error inserting user %s: %v", user.LineUserID, err)
			errors++
			continue
		}

		count++
		if count%10 == 0 {
			log.Printf("  Migrated %d users...", count)
		}
	}

	log.Printf("✓ Migrated %d users (%d errors)", count, errors)
}

func insertUser(db *sql.DB, user *User) error {
	query := `
		INSERT INTO users (line_user_id, line_display_name, picture_url, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (line_user_id) DO UPDATE SET
			line_display_name = EXCLUDED.line_display_name,
			picture_url = EXCLUDED.picture_url,
			role = EXCLUDED.role
	`
	_, err := db.Exec(query, user.LineUserID, user.LineDisplayName, user.PictureURL, user.Role, user.CreatedAt)
	return err
}

func migrateEvents(ctx context.Context, fs *firestore.Client, pg *sql.DB) {
	log.Println("\n--- Migrating Events ---")

	iter := fs.Collection("events").Documents(ctx)
	count := 0
	errors := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading event: %v", err)
			errors++
			continue
		}

		var event Event
		if err := doc.DataTo(&event); err != nil {
			log.Printf("Error parsing event %s: %v", doc.Ref.ID, err)
			errors++
			continue
		}

		// Use document ID as EventID if not set
		if event.EventID == "" {
			event.EventID = doc.Ref.ID
		}

		err = insertEvent(pg, &event)
		if err != nil {
			log.Printf("Error inserting event %s: %v", event.EventID, err)
			errors++
			continue
		}

		count++
	}

	log.Printf("✓ Migrated %d events (%d errors)", count, errors)
}

func insertEvent(db *sql.DB, event *Event) error {
	configJSON, err := json.Marshal(event.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO events (event_id, type, title, is_active, created_by, created_at, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id) DO UPDATE SET
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			is_active = EXCLUDED.is_active,
			config = EXCLUDED.config
	`
	_, err = db.Exec(query, event.EventID, event.Type, event.Title, event.IsActive, event.CreatedBy, event.CreatedAt, configJSON)
	return err
}

func migrateInteractions(ctx context.Context, fs *firestore.Client, pg *sql.DB) {
	log.Println("\n--- Migrating Interactions ---")

	// Get all events first
	eventsIter := fs.Collection("events").Documents(ctx)
	totalCount := 0
	totalErrors := 0

	for {
		eventDoc, err := eventsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error reading event: %v", err)
			continue
		}

		eventID := eventDoc.Ref.ID

		// Get records subcollection for this event
		recordsIter := eventDoc.Ref.Collection("records").Documents(ctx)
		eventCount := 0

		for {
			doc, err := recordsIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Error reading interaction: %v", err)
				totalErrors++
				continue
			}

			var interaction Interaction
			if err := doc.DataTo(&interaction); err != nil {
				log.Printf("Error parsing interaction %s: %v", doc.Ref.ID, err)
				totalErrors++
				continue
			}

			interaction.ID = doc.Ref.ID

			err = insertInteraction(pg, eventID, &interaction)
			if err != nil {
				log.Printf("Error inserting interaction %s: %v", interaction.ID, err)
				totalErrors++
				continue
			}

			eventCount++
			totalCount++
		}

		if eventCount > 0 {
			log.Printf("  Event %s: %d interactions", eventID[:8], eventCount)
		}
	}

	log.Printf("✓ Migrated %d interactions (%d errors)", totalCount, totalErrors)
}

func insertInteraction(db *sql.DB, eventID string, interaction *Interaction) error {
	payload := InteractionPayload{
		SelectedOptions: interaction.SelectedOptions,
		Count:           interaction.Count,
		Note:            interaction.Note,
		Content:         interaction.Content,
		ClapCount:       interaction.ClapCount,
		Reactions:       interaction.Reactions,
	}

	if !interaction.CancelledAt.IsZero() {
		payload.CancelledAt = interaction.CancelledAt.Format(time.RFC3339)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO interactions (id, event_id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_display_name = EXCLUDED.user_display_name,
			user_picture_url = EXCLUDED.user_picture_url,
			status = EXCLUDED.status,
			payload = EXCLUDED.payload
	`
	_, err = db.Exec(query, interaction.ID, eventID, interaction.UserID, interaction.Type,
		interaction.UserDisplayName, interaction.UserPictureUrl, interaction.Status, interaction.Timestamp, payloadJSON)
	return err
}

func verifyMigration(ctx context.Context, fs *firestore.Client, pg *sql.DB) {
	log.Println("\n--- Verifying Migration ---")

	// Count Firestore
	fsUsers := countFirestoreCollection(ctx, fs, "users")
	fsEvents := countFirestoreCollection(ctx, fs, "events")
	fsInteractions := countFirestoreInteractions(ctx, fs)

	// Count PostgreSQL
	pgUsers := countPostgresTable(pg, "users")
	pgEvents := countPostgresTable(pg, "events")
	pgInteractions := countPostgresTable(pg, "interactions")

	fmt.Println("\n┌─────────────────┬───────────┬────────────┬─────────┐")
	fmt.Println("│ Table           │ Firestore │ PostgreSQL │ Status  │")
	fmt.Println("├─────────────────┼───────────┼────────────┼─────────┤")
	fmt.Printf("│ users           │ %9d │ %10d │ %s │\n", fsUsers, pgUsers, getStatus(fsUsers, pgUsers))
	fmt.Printf("│ events          │ %9d │ %10d │ %s │\n", fsEvents, pgEvents, getStatus(fsEvents, pgEvents))
	fmt.Printf("│ interactions    │ %9d │ %10d │ %s │\n", fsInteractions, pgInteractions, getStatus(fsInteractions, pgInteractions))
	fmt.Println("└─────────────────┴───────────┴────────────┴─────────┘")
}

func getStatus(fs, pg int) string {
	if fs == pg {
		return "  ✓    "
	}
	return "  ✗    "
}

func dryRun(ctx context.Context, fs *firestore.Client) {
	log.Println("\n--- Dry Run (no data will be written) ---")

	users := countFirestoreCollection(ctx, fs, "users")
	events := countFirestoreCollection(ctx, fs, "events")
	interactions := countFirestoreInteractions(ctx, fs)

	fmt.Println("\nData to migrate:")
	fmt.Printf("  Users:        %d\n", users)
	fmt.Printf("  Events:       %d\n", events)
	fmt.Printf("  Interactions: %d\n", interactions)
	fmt.Printf("  Total:        %d records\n", users+events+interactions)
}

func countFirestoreCollection(ctx context.Context, fs *firestore.Client, collection string) int {
	iter := fs.Collection(collection).Documents(ctx)
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		count++
	}
	return count
}

func countFirestoreInteractions(ctx context.Context, fs *firestore.Client) int {
	eventsIter := fs.Collection("events").Documents(ctx)
	total := 0

	for {
		eventDoc, err := eventsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		recordsIter := eventDoc.Ref.Collection("records").Documents(ctx)
		for {
			_, err := recordsIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				continue
			}
			total++
		}
	}
	return total
}

func countPostgresTable(db *sql.DB, table string) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
