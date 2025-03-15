package discord

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/azlyth/irlcord/pkg/models"
	"github.com/bwmarrin/discordgo"
)

// Colors for embeds
const (
	ColorSuccess = 0x43B581
	ColorError   = 0xF04747
	ColorInfo    = 0x7289DA
	ColorWarning = 0xFAA61A
)

// SendMessage sends a message to a channel
func SendMessage(s *discordgo.Session, channelID, content string) (*discordgo.Message, error) {
	return s.ChannelMessageSend(channelID, content)
}

// SendEmbed sends an embed to a channel
func SendEmbed(s *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return s.ChannelMessageSendEmbed(channelID, embed)
}

// SendErrorMessage sends an error message to a channel
func SendErrorMessage(s *discordgo.Session, channelID, content string) (*discordgo.Message, error) {
	embed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: content,
		Color:       ColorError,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return SendEmbed(s, channelID, embed)
}

// SendSuccessMessage sends a success message to a channel
func SendSuccessMessage(s *discordgo.Session, channelID, content string) (*discordgo.Message, error) {
	embed := &discordgo.MessageEmbed{
		Title:       "Success",
		Description: content,
		Color:       ColorSuccess,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return SendEmbed(s, channelID, embed)
}

// CreateEventEmbed creates an embed for an event
func CreateEventEmbed(event *models.Event, attendees []*models.EventAttendee, groupSingular string) *discordgo.MessageEmbed {
	// Format date and time
	dateStr := event.DateTime.Format("Monday, January 2, 2006")
	timeStr := event.DateTime.Format("3:04 PM")

	// Create description
	description := fmt.Sprintf("**Date:** %s\n**Time:** %s\n", dateStr, timeStr)

	if event.LocationName != "" {
		description += fmt.Sprintf("**Location:** %s\n", event.LocationName)
	}

	if event.LocationAddress != "" {
		description += fmt.Sprintf("**Address:** %s\n", event.LocationAddress)
	}

	if event.Description != "" {
		description += fmt.Sprintf("\n%s\n", event.Description)
	}

	// Create fields for attendees
	var fields []*discordgo.MessageEmbedField

	if len(attendees) > 0 {
		attending := []*models.EventAttendee{}
		waitlist := []*models.EventAttendee{}
		declined := []*models.EventAttendee{}

		for _, attendee := range attendees {
			switch attendee.RSVPStatus {
			case string(models.RSVPStatusAttending):
				attending = append(attending, attendee)
			case string(models.RSVPStatusWaitlist):
				waitlist = append(waitlist, attendee)
			case string(models.RSVPStatusDeclined):
				declined = append(declined, attendee)
			}
		}

		if len(attending) > 0 {
			attendeeNames := []string{}
			for _, attendee := range attending {
				attendeeNames = append(attendeeNames, fmt.Sprintf("<@%s>", attendee.UserID))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Attending (%d)", len(attending)),
				Value:  strings.Join(attendeeNames, "\n"),
				Inline: true,
			})
		}

		if len(waitlist) > 0 {
			waitlistNames := []string{}
			for _, attendee := range waitlist {
				waitlistNames = append(waitlistNames, fmt.Sprintf("<@%s>", attendee.UserID))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Waitlist (%d)", len(waitlist)),
				Value:  strings.Join(waitlistNames, "\n"),
				Inline: true,
			})
		}

		if len(declined) > 0 {
			declinedNames := []string{}
			for _, attendee := range declined {
				declinedNames = append(declinedNames, fmt.Sprintf("<@%s>", attendee.UserID))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Declined (%d)", len(declined)),
				Value:  strings.Join(declinedNames, "\n"),
				Inline: true,
			})
		}
	}

	// Status indicator
	statusEmoji := "🟢" // Default: Approved
	if event.Status == string(models.EventStatusPending) {
		statusEmoji = "🟠"
	} else if event.Status == string(models.EventStatusRejected) {
		statusEmoji = "🔴"
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", statusEmoji, event.Name),
		Description: description,
		Color:       ColorInfo,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Event ID: %d • Host: <@%s>", event.EventID, event.HostID),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return embed
}

// CreateGroupEmbed creates an embed for a group
func CreateGroupEmbed(group *models.Group, members []*models.GroupMember, terminology map[string]string) *discordgo.MessageEmbed {
	// Create description
	description := group.Description
	if description == "" {
		description = "No description provided."
	}

	// Add group settings
	settings := []string{}

	if group.IsOpen {
		settings = append(settings, "Open Group: Yes")
	} else {
		settings = append(settings, "Open Group: No")
	}

	if group.NewMembersCanCreateEvents {
		settings = append(settings, "New Members Can Create Events: Yes")
	} else {
		settings = append(settings, "New Members Can Create Events: No")
	}

	if group.EventApprovalMode != "" {
		settings = append(settings, fmt.Sprintf("Event Approval Mode: %s", strings.Title(group.EventApprovalMode)))
	}

	if group.EventAttendeeManagementMode != "" {
		settings = append(settings, fmt.Sprintf("Attendee Management: %s", strings.Title(group.EventAttendeeManagementMode)))
	}

	if group.ContributorEventsRequired > 0 {
		settings = append(settings, fmt.Sprintf("Events Required for Contributor: %d", group.ContributorEventsRequired))
	}

	if len(settings) > 0 {
		description += "\n\n**Settings:**\n" + strings.Join(settings, "\n")
	}

	// Create fields for members
	var fields []*discordgo.MessageEmbedField

	if len(members) > 0 {
		leaders := []*models.GroupMember{}
		regularMembers := []*models.GroupMember{}

		for _, member := range members {
			if member.IsLeader {
				leaders = append(leaders, member)
			} else {
				regularMembers = append(regularMembers, member)
			}
		}

		if len(leaders) > 0 {
			leaderNames := []string{}
			for _, leader := range leaders {
				leaderNames = append(leaderNames, fmt.Sprintf("<@%s>", leader.UserID))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Leaders",
				Value:  strings.Join(leaderNames, "\n"),
				Inline: true,
			})
		}

		if len(regularMembers) > 0 {
			memberNames := []string{}
			for _, member := range regularMembers[:min(10, len(regularMembers))] {
				memberNames = append(memberNames, fmt.Sprintf("<@%s>", member.UserID))
			}
			if len(regularMembers) > 10 {
				memberNames = append(memberNames, fmt.Sprintf("... and %d more", len(regularMembers)-10))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Members (%d)", len(regularMembers)),
				Value:  strings.Join(memberNames, "\n"),
				Inline: true,
			})
		}
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       group.Name,
		Description: description,
		Color:       ColorInfo,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Group ID: %d • Created: %s", group.GroupID, group.CreatedAt.Format(time.RFC3339)),
		},
	}

	return embed
}

// GetOrCreateThread gets an existing thread by name or creates a new one
func GetOrCreateThread(s *discordgo.Session, channelID, name string, message *discordgo.Message) (*discordgo.Channel, error) {
	// Get all threads in the channel
	threads, err := s.ThreadsActive(s.State.User.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active threads: %w", err)
	}

	// Try to find an existing thread with the same name
	for _, thread := range threads.Threads {
		if thread.ParentID == channelID && thread.Name == name && !thread.ThreadMetadata.Archived {
			return thread, nil
		}
	}

	// Create a new thread
	var thread *discordgo.Channel
	if message != nil {
		// Create a thread from a message
		thread, err = s.MessageThreadStartComplex(channelID, message.ID, &discordgo.ThreadStart{
			Name:                name,
			AutoArchiveDuration: 1440, // 1 day in minutes
		})
	} else {
		// Create a thread without a message
		thread, err = s.ThreadStartComplex(channelID, &discordgo.ThreadStart{
			Name:                name,
			Type:                discordgo.ChannelTypeGuildPublicThread,
			AutoArchiveDuration: 1440, // 1 day in minutes
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return thread, nil
}

// ParseCommandArgs parses command arguments from a message content
func ParseCommandArgs(content string) map[string]string {
	// Remove command prefix and command name
	parts := strings.SplitN(content, " ", 2)
	if len(parts) < 2 {
		return map[string]string{}
	}

	argsText := parts[1]
	args := map[string]string{}

	// Match quoted values first
	quotedPattern := regexp.MustCompile(`(\w+)="([^"]*)"`)
	for _, match := range quotedPattern.FindAllStringSubmatch(argsText, -1) {
		if len(match) >= 3 {
			key := strings.ToLower(match[1])
			value := match[2]
			args[key] = value
			argsText = strings.Replace(argsText, match[0], "", 1)
		}
	}

	// Then match unquoted values
	unquotedPattern := regexp.MustCompile(`(\w+)=(\S+)`)
	for _, match := range unquotedPattern.FindAllStringSubmatch(argsText, -1) {
		if len(match) >= 3 {
			key := strings.ToLower(match[1])
			value := match[2]
			args[key] = value
		}
	}

	return args
}

// ExtractUserID extracts a user ID from a mention
func ExtractUserID(mention string) string {
	// Remove <@, <@!, and > from the mention
	mention = strings.TrimPrefix(mention, "<@")
	mention = strings.TrimPrefix(mention, "!")
	mention = strings.TrimSuffix(mention, ">")
	return mention
}

// ParseInt parses a string to an int with a default value
func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Error parsing int: %v", err)
		return defaultValue
	}
	return i
}

// ParseBool parses a string to a bool with a default value
func ParseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Printf("Error parsing bool: %v", err)
		return defaultValue
	}
	return b
}

// ParseTime parses a date and time string to a time.Time
func ParseTime(dateStr, timeStr string) (time.Time, error) {
	// Parse date and time
	dateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)
	return time.Parse("2006-01-02 15:04", dateTimeStr)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 