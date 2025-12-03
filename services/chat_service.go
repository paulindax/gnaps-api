package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ChatService struct {
	apiKey  string
	apiURL  string
	enabled bool
}

type ChatRequest struct {
	Message string      `json:"message"`
	Context ChatContext `json:"context"`
}

type ChatContext struct {
	UserRole string `json:"user_role"`
	UserName string `json:"user_name"`
}

type ChatResponse struct {
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// Claude API structures
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []ClaudeMessage `json:"messages"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewChatService() *ChatService {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	return &ChatService{
		apiKey:  apiKey,
		apiURL:  "https://api.anthropic.com/v1/messages",
		enabled: apiKey != "",
	}
}

func (s *ChatService) GetResponse(request ChatRequest) (*ChatResponse, error) {
	// If AI is enabled, use Claude API
	if s.enabled {
		return s.getAIResponse(request)
	}

	// Fall back to local knowledge base
	return s.getLocalResponse(request)
}

func (s *ChatService) getAIResponse(request ChatRequest) (*ChatResponse, error) {
	systemPrompt := s.buildSystemPrompt(request.Context)

	claudeReq := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []ClaudeMessage{
			{Role: "user", Content: request.Message},
		},
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return s.getLocalResponse(request)
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return s.getLocalResponse(request)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return s.getLocalResponse(request)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.getLocalResponse(request)
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return s.getLocalResponse(request)
	}

	if claudeResp.Error != nil {
		return s.getLocalResponse(request)
	}

	if len(claudeResp.Content) > 0 {
		return &ChatResponse{
			Message: claudeResp.Content[0].Text,
		}, nil
	}

	return s.getLocalResponse(request)
}

func (s *ChatService) buildSystemPrompt(context ChatContext) string {
	roleDescription := s.getRoleDescription(context.UserRole)

	return fmt.Sprintf(`You are Adesua360, a friendly and helpful AI assistant for the GNAPS (Ghana National Association of Private Schools) management application. Your name means "learning" in Akan.

ABOUT GNAPS:
GNAPS is a comprehensive school management system that helps manage private schools across Ghana. The system is organized hierarchically:
- National Level: Overall administration
- Regions: Geographic divisions of Ghana (e.g., Greater Accra, Ashanti, etc.)
- Zones: Sub-divisions within regions
- Groups: Categories of schools
- Schools: Individual educational institutions

KEY FEATURES YOU CAN HELP WITH:
1. **Schools Management**: Adding, viewing, editing schools. Schools have details like name, address, contact info, and belong to a Zone and Group.

2. **Regions & Zones**: Administrative divisions. Regions contain Zones, and Zones contain Schools.

3. **Executives**: Association leaders at national, regional, and zonal levels:
   - National Admin: Manages entire system
   - Regional Admin: Manages a specific region
   - Zonal Admin: Manages a specific zone

4. **Payments & Finance**: Track school dues, payments, and financial records.

5. **News & Events**: Create and manage announcements, news articles, and events.

6. **Document Vault**: Create form templates and manage document submissions from schools.

USER CONTEXT:
- User Role: %s
- User Name: %s
%s

GUIDELINES:
- Be friendly, helpful, and concise
- Provide step-by-step instructions when explaining how to do something
- Use bullet points and formatting for clarity
- If you don't know something specific, suggest contacting support
- Focus on helping the user navigate and use the GNAPS application
- Keep responses focused on GNAPS-related topics
- Use **bold** for emphasis on important terms or buttons
- If the user asks about something outside your knowledge, politely redirect to GNAPS topics

Always respond in a helpful, clear manner as if you're a knowledgeable colleague who knows the system well.`, context.UserRole, context.UserName, roleDescription)
}

func (s *ChatService) getRoleDescription(role string) string {
	descriptions := map[string]string{
		"system_admin":   "This user is a System Administrator with full access to all features including system configuration.",
		"national_admin": "This user is a National Administrator who can manage all regions, zones, schools, and executives.",
		"region_admin":   "This user is a Regional Administrator who can manage zones and schools within their assigned region.",
		"zone_admin":     "This user is a Zonal Administrator who can manage schools within their assigned zone.",
		"school_admin":   "This user is a School Administrator who can manage their own school's information and submissions.",
	}

	if desc, ok := descriptions[role]; ok {
		return desc
	}
	return "This user has standard access to the system."
}

func (s *ChatService) getLocalResponse(request ChatRequest) (*ChatResponse, error) {
	query := strings.ToLower(request.Message)

	// School-related queries
	if containsAny(query, []string{"school", "add school", "create school", "new school"}) {
		if containsAny(query, []string{"add", "create", "new"}) {
			return &ChatResponse{
				Message: `To add a new school:

1. Go to **Schools** in the sidebar menu
2. Click the **"Add School"** button
3. Fill in the school details:
   - School name
   - Address and location
   - Contact information
   - Select the Zone
   - Select the Group/Category
4. Click **"Save"** to create the school

Would you like more details about any specific field?`,
			}, nil
		}
		return &ChatResponse{
			Message: `To view and manage schools:

1. Click on **Schools** in the sidebar
2. Use the **filters** to narrow down by:
   - Region
   - Zone
   - Group
3. Use the **search bar** to find specific schools
4. Click on any school to view or edit its details

You can also export the school list for reporting.`,
		}, nil
	}

	// Payment queries
	if containsAny(query, []string{"payment", "pay", "dues", "fee", "finance"}) {
		return &ChatResponse{
			Message: `To manage payments:

1. Navigate to **Payments** in the sidebar
2. View the payment dashboard for an overview
3. Use filters to see payments by:
   - Date range
   - Region/Zone
   - Payment status
4. Click on any payment record for full details

To track a specific school's payment:
- Go to the school's profile
- Check the **Payments** tab

Need help with a specific payment issue?`,
		}, nil
	}

	// Executive queries
	if containsAny(query, []string{"executive", "admin", "role"}) {
		return &ChatResponse{
			Message: `To manage executives:

1. Go to **Executives** in the sidebar
2. Click **"Add Executive"** to create new
3. Fill in personal details (name, email, phone)
4. Select the **Role**:
   - **National Admin**: Full system access
   - **Regional Admin**: Manages a region
   - **Zonal Admin**: Manages a zone
5. Assign to the appropriate Region/Zone
6. Click **"Save"**

Existing executives can be edited or deactivated from their profile.`,
		}, nil
	}

	// Region/Zone queries
	if containsAny(query, []string{"region", "zone"}) {
		return &ChatResponse{
			Message: `To manage regions and zones:

**Regions:**
1. Go to **Settings** > **Regions**
2. Click **"Add Region"** to create new
3. Enter region name and code
4. Assign administrators

**Zones:**
1. Go to **Settings** > **Zones**
2. Select the parent **Region** first
3. Click **"Add Zone"** to create new
4. Enter zone details
5. Assign a Zonal Admin

The hierarchy is: National > Regions > Zones > Schools`,
		}, nil
	}

	// News/Events queries
	if containsAny(query, []string{"news", "article", "announcement", "event"}) {
		return &ChatResponse{
			Message: `To manage news and events:

**News Articles:**
1. Go to **News** in the sidebar
2. Click **"Create Article"**
3. Add title, content, and featured image
4. Set publish date
5. Click **"Publish"** or save as draft

**Events:**
1. Go to **Events** in the sidebar
2. Click **"Create Event"**
3. Fill in event details (title, date, location)
4. Add event image
5. Publish when ready

Both can be scheduled for future dates.`,
		}, nil
	}

	// Document queries
	if containsAny(query, []string{"document", "template", "form", "submission"}) {
		return &ChatResponse{
			Message: `To work with the Document Vault:

**Creating Templates:**
1. Go to **Document Vault** > **Templates**
2. Click **"Create Template"**
3. Use the visual builder to add fields:
   - Text inputs
   - Checkboxes
   - Signatures
   - Date fields
4. Save and publish

**Managing Submissions:**
1. Go to **Document Vault** > **Submissions**
2. View forms submitted by schools
3. Filter by template, school, or status
4. Download or review submissions`,
		}, nil
	}

	// Dashboard queries
	if containsAny(query, []string{"dashboard", "overview", "statistics", "stats"}) {
		return &ChatResponse{
			Message: `Your **Dashboard** provides:

- **Quick Stats**: Total schools, zones, regions at a glance
- **Payment Overview**: Recent payments and trends
- **Activity Feed**: Latest updates across the system
- **Charts**: Visual representation of key metrics

Click on any stat card to drill down into details.

The dashboard is customized based on your role and shows data relevant to your jurisdiction.`,
		}, nil
	}

	// Help/Support queries
	if containsAny(query, []string{"help", "support", "contact", "issue"}) {
		return &ChatResponse{
			Message: `For support:

1. **In-app Help**: I'm here to help! Ask me anything
2. **Documentation**: Check the Help section in Settings
3. **Email**: Contact support@gnaps.org
4. **Phone**: Reach your regional administrator

Common troubleshooting:
- Clear browser cache for display issues
- Check internet connection for loading errors
- Ensure you have the correct permissions

What specific issue can I help you with?`,
		}, nil
	}

	// Features/Capabilities queries
	if containsAny(query, []string{"what can you", "help me with", "features", "capable"}) {
		return &ChatResponse{
			Message: `I can help you with:

**Navigation & Usage**
- Finding features in the app
- Step-by-step instructions
- Understanding different sections

**School Management**
- Adding and viewing schools
- Managing school information
- Understanding the hierarchy

**Administration**
- Managing regions and zones
- Executive role assignments
- User permissions

**Finance**
- Payment tracking
- Financial reports
- Billing information

**Content**
- News and announcements
- Events management
- Document templates

Just ask me anything about GNAPS!`,
		}, nil
	}

	// Default response
	return &ChatResponse{
		Message: fmt.Sprintf(`I understand you're asking about "%s".

Here are some things I can help you with:

- **Schools**: Adding, viewing, and managing schools
- **Payments**: Tracking dues and financial records
- **Regions/Zones**: Administrative divisions
- **Executives**: Role assignments and management
- **Events & News**: Content management
- **Documents**: Form templates and submissions

Could you be more specific about what you'd like to do? I'm here to help!`, request.Message),
	}, nil
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
