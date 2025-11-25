# 🚨 Report System Guide

## Overview

The SkillMatch platform includes a **comprehensive user reporting system** for maintaining platform safety and integrity. The system allows users to report inappropriate behavior, content, or violations of platform policies.

**Key Features:**
- Multiple report types
- Anti-spam protection (prevent duplicate reports)
- Admin moderation workflow
- Status tracking (pending → reviewing → resolved/dismissed)
- Audit trail with timestamps

---

## Database Schema

### `reports` Table

```sql
CREATE TABLE reports (
    id SERIAL PRIMARY KEY,
    reporter_id INTEGER REFERENCES users(user_id),  -- User making the report
    reported_user_id INTEGER REFERENCES users(user_id),  -- User being reported
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    admin_notes TEXT,  -- Notes added by admin during review
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_reason CHECK (reason IN (
        'harassment',
        'inappropriate_content',
        'fake_profile',
        'scam',
        'violence_threat',
        'underage',
        'spam',
        'other'
    )),
    CONSTRAINT check_status CHECK (status IN (
        'pending',
        'reviewing',
        'resolved',
        'dismissed'
    )),
    CONSTRAINT check_not_self_report CHECK (reporter_id != reported_user_id)
);
```

**Key Points:**
- Tracks who reported whom
- Prevents self-reporting
- Status workflow for admin moderation
- Auto-updated timestamp trigger
- Admin notes for documentation

---

## Report Types

| Reason | Description | Example |
|--------|-------------|---------|
| `harassment` | Abusive or threatening behavior | Repeated unwanted messages, threats |
| `inappropriate_content` | Sexual content, nudity violations | Profile photos violating guidelines |
| `fake_profile` | Impersonation or false identity | Using someone else's photos/identity |
| `scam` | Fraudulent activity | Requesting money outside platform |
| `violence_threat` | Threats of physical harm | Direct threats of violence |
| `underage` | Suspected minor | Profile appears to be under 18 |
| `spam` | Unsolicited commercial content | Mass messages promoting services |
| `other` | Other violations | Anything not covered above |

---

## API Endpoints

### User Endpoints

#### 1. Create a Report

**POST** `/reports`

Allows a user to report another user.

**Request Body:**
```json
{
  "reported_user_id": 42,
  "reason": "harassment",
  "description": "This user has been sending me threatening messages repeatedly after I declined their booking request."
}
```

**Validation:**
- `reported_user_id` (required): Must be a valid user ID
- `reason` (required): Must be one of the valid report types
- `description` (optional): Detailed explanation (max 2000 characters)

**Anti-Spam Protection:**
- Prevents duplicate reports within 24 hours
- Returns error if user already reported this person recently

**Response (Success):**
```json
{
  "message": "Report submitted successfully",
  "report": {
    "id": 15,
    "reporter_id": 5,
    "reported_user_id": 42,
    "reason": "harassment",
    "description": "This user has been sending me threatening messages...",
    "status": "pending",
    "created_at": "2025-11-13T10:00:00Z"
  }
}
```

**Response (Duplicate):**
```json
{
  "error": "You have already reported this user within the last 24 hours"
}
```

---

#### 2. Get My Reports

**GET** `/reports/my`

Returns all reports created by the authenticated user.

**Response:**
```json
{
  "reports": [
    {
      "id": 15,
      "reporter_id": 5,
      "reported_user_id": 42,
      "reported_user_name": "John Doe",
      "reason": "harassment",
      "description": "This user has been sending me threatening messages...",
      "status": "reviewing",
      "created_at": "2025-11-13T10:00:00Z",
      "updated_at": "2025-11-13T11:00:00Z"
    }
  ]
}
```

**Use Cases:**
- View status of submitted reports
- Track reported users
- Check if report is under review

---

### Admin Endpoints

All admin endpoints require **admin authentication** via `adminAuthMiddleware`.

#### 3. Get All Reports

**GET** `/admin/reports?status=pending&limit=50&offset=0`

Returns all reports in the system (admin only).

**Query Parameters:**
- `status` (optional): Filter by status (pending, reviewing, resolved, dismissed)
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:**
```json
{
  "reports": [
    {
      "id": 15,
      "reporter_id": 5,
      "reporter_name": "Alice Smith",
      "reported_user_id": 42,
      "reported_user_name": "John Doe",
      "reason": "harassment",
      "description": "This user has been sending me threatening messages...",
      "status": "pending",
      "admin_notes": null,
      "created_at": "2025-11-13T10:00:00Z",
      "updated_at": "2025-11-13T10:00:00Z"
    },
    {
      "id": 14,
      "reporter_id": 8,
      "reporter_name": "Bob Johnson",
      "reported_user_id": 42,
      "reported_user_name": "John Doe",
      "reason": "fake_profile",
      "description": "Profile photos are stolen from Instagram",
      "status": "reviewing",
      "admin_notes": "Investigating social media profiles",
      "created_at": "2025-11-12T15:00:00Z",
      "updated_at": "2025-11-13T09:00:00Z"
    }
  ],
  "total": 2
}
```

---

#### 4. Update Report Status

**PATCH** `/admin/reports/:id`

Updates the status and admin notes for a report (admin only).

**Request Body:**
```json
{
  "status": "resolved",
  "admin_notes": "User has been warned and profile content removed. Monitoring for repeated violations."
}
```

**Valid Status Transitions:**
- `pending` → `reviewing`
- `reviewing` → `resolved` or `dismissed`
- `pending` → `resolved` or `dismissed` (direct resolution)

**Response:**
```json
{
  "message": "Report updated successfully",
  "report": {
    "id": 15,
    "status": "resolved",
    "admin_notes": "User has been warned and profile content removed. Monitoring for repeated violations.",
    "updated_at": "2025-11-13T12:00:00Z"
  }
}
```

---

#### 5. Delete Report

**DELETE** `/admin/reports/:id`

Permanently deletes a report (admin only).

**Use Cases:**
- Remove duplicate reports
- Clean up test reports
- Remove reports created in error

**Response:**
```json
{
  "message": "Report deleted successfully"
}
```

---

## Admin Moderation Workflow

### Step 1: Review Pending Reports

```bash
# Get all pending reports
curl -X GET "http://localhost:8080/admin/reports?status=pending" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Step 2: Investigate

1. Review report description
2. Check reported user's profile
3. Review chat history (if harassment)
4. Check booking history
5. Look for pattern of violations

### Step 3: Take Action

**For legitimate violations:**

```bash
# Mark as reviewing
curl -X PATCH http://localhost:8080/admin/reports/15 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "reviewing",
    "admin_notes": "Investigating chat history and user profile"
  }'

# After investigation, resolve
curl -X PATCH http://localhost:8080/admin/reports/15 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved",
    "admin_notes": "User warned. Profile content removed. Account suspended for 7 days."
  }'
```

**For false/spam reports:**

```bash
# Dismiss the report
curl -X PATCH http://localhost:8080/admin/reports/15 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "dismissed",
    "admin_notes": "No evidence of violation found. Report appears to be made in bad faith."
  }'
```

---

## Admin Dashboard Implementation

### React Admin Component

```typescript
import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface Report {
  id: number;
  reporter_id: number;
  reporter_name: string;
  reported_user_id: number;
  reported_user_name: string;
  reason: string;
  description: string;
  status: string;
  admin_notes?: string;
  created_at: string;
  updated_at: string;
}

export const AdminReportsPanel: React.FC = () => {
  const [reports, setReports] = useState<Report[]>([]);
  const [filter, setFilter] = useState<string>('pending');
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [adminNotes, setAdminNotes] = useState('');

  useEffect(() => {
    fetchReports();
  }, [filter]);

  const fetchReports = async () => {
    const response = await axios.get(`/admin/reports?status=${filter}`);
    setReports(response.data.reports);
  };

  const updateReportStatus = async (reportId: number, newStatus: string) => {
    await axios.patch(`/admin/reports/${reportId}`, {
      status: newStatus,
      admin_notes: adminNotes
    });
    fetchReports();
    setSelectedReport(null);
    setAdminNotes('');
  };

  const deleteReport = async (reportId: number) => {
    if (confirm('Are you sure you want to delete this report?')) {
      await axios.delete(`/admin/reports/${reportId}`);
      fetchReports();
    }
  };

  return (
    <div className="admin-reports-panel">
      <h1>Report Management</h1>

      {/* Filters */}
      <div className="filters">
        <button 
          className={filter === 'pending' ? 'active' : ''}
          onClick={() => setFilter('pending')}
        >
          Pending
        </button>
        <button 
          className={filter === 'reviewing' ? 'active' : ''}
          onClick={() => setFilter('reviewing')}
        >
          Reviewing
        </button>
        <button 
          className={filter === 'resolved' ? 'active' : ''}
          onClick={() => setFilter('resolved')}
        >
          Resolved
        </button>
        <button 
          className={filter === 'dismissed' ? 'active' : ''}
          onClick={() => setFilter('dismissed')}
        >
          Dismissed
        </button>
      </div>

      {/* Reports Table */}
      <table className="reports-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Reporter</th>
            <th>Reported User</th>
            <th>Reason</th>
            <th>Created</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {reports.map(report => (
            <tr key={report.id} className={`status-${report.status}`}>
              <td>#{report.id}</td>
              <td>
                <a href={`/admin/users/${report.reporter_id}`}>
                  {report.reporter_name}
                </a>
              </td>
              <td>
                <a href={`/admin/users/${report.reported_user_id}`}>
                  {report.reported_user_name}
                </a>
              </td>
              <td>
                <span className={`reason-badge ${report.reason}`}>
                  {formatReason(report.reason)}
                </span>
              </td>
              <td>{formatDate(report.created_at)}</td>
              <td>
                <span className={`status-badge ${report.status}`}>
                  {report.status}
                </span>
              </td>
              <td>
                <button onClick={() => setSelectedReport(report)}>
                  View Details
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Report Details Modal */}
      {selectedReport && (
        <div className="modal">
          <div className="modal-content">
            <h2>Report Details - #{selectedReport.id}</h2>
            
            <div className="report-info">
              <p><strong>Reporter:</strong> {selectedReport.reporter_name}</p>
              <p><strong>Reported User:</strong> {selectedReport.reported_user_name}</p>
              <p><strong>Reason:</strong> {formatReason(selectedReport.reason)}</p>
              <p><strong>Status:</strong> {selectedReport.status}</p>
              <p><strong>Created:</strong> {formatDate(selectedReport.created_at)}</p>
            </div>

            <div className="description">
              <h3>Description</h3>
              <p>{selectedReport.description}</p>
            </div>

            {selectedReport.admin_notes && (
              <div className="admin-notes-display">
                <h3>Previous Admin Notes</h3>
                <p>{selectedReport.admin_notes}</p>
              </div>
            )}

            <div className="admin-notes-input">
              <h3>Admin Notes</h3>
              <textarea
                value={adminNotes}
                onChange={(e) => setAdminNotes(e.target.value)}
                placeholder="Add notes about your investigation and actions taken..."
                rows={4}
              />
            </div>

            <div className="actions">
              {selectedReport.status === 'pending' && (
                <>
                  <button 
                    className="btn-primary"
                    onClick={() => updateReportStatus(selectedReport.id, 'reviewing')}
                  >
                    Start Reviewing
                  </button>
                  <button 
                    className="btn-success"
                    onClick={() => updateReportStatus(selectedReport.id, 'resolved')}
                  >
                    Resolve Immediately
                  </button>
                  <button 
                    className="btn-warning"
                    onClick={() => updateReportStatus(selectedReport.id, 'dismissed')}
                  >
                    Dismiss
                  </button>
                </>
              )}

              {selectedReport.status === 'reviewing' && (
                <>
                  <button 
                    className="btn-success"
                    onClick={() => updateReportStatus(selectedReport.id, 'resolved')}
                  >
                    Mark as Resolved
                  </button>
                  <button 
                    className="btn-warning"
                    onClick={() => updateReportStatus(selectedReport.id, 'dismissed')}
                  >
                    Dismiss
                  </button>
                </>
              )}

              <button 
                className="btn-danger"
                onClick={() => deleteReport(selectedReport.id)}
              >
                Delete Report
              </button>

              <button onClick={() => setSelectedReport(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

function formatReason(reason: string): string {
  return reason.split('_').map(word => 
    word.charAt(0).toUpperCase() + word.slice(1)
  ).join(' ');
}

function formatDate(timestamp: string): string {
  return new Date(timestamp).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}
```

---

## User Report Flow

### User Interface Component

```typescript
import React, { useState } from 'react';
import axios from 'axios';

interface ReportModalProps {
  userId: number;
  userName: string;
  onClose: () => void;
}

export const ReportUserModal: React.FC<ReportModalProps> = ({ 
  userId, 
  userName, 
  onClose 
}) => {
  const [reason, setReason] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const reportReasons = [
    { value: 'harassment', label: 'Harassment or Threats' },
    { value: 'inappropriate_content', label: 'Inappropriate Content' },
    { value: 'fake_profile', label: 'Fake Profile or Impersonation' },
    { value: 'scam', label: 'Scam or Fraud' },
    { value: 'violence_threat', label: 'Threat of Violence' },
    { value: 'underage', label: 'Underage User' },
    { value: 'spam', label: 'Spam' },
    { value: 'other', label: 'Other' }
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await axios.post('/reports', {
        reported_user_id: userId,
        reason,
        description
      });

      alert('Report submitted successfully. Our team will review it shortly.');
      onClose();
    } catch (err: any) {
      if (err.response?.status === 400) {
        setError(err.response.data.error);
      } else {
        setError('Failed to submit report. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal">
      <div className="modal-content">
        <h2>Report User: {userName}</h2>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Reason for Report *</label>
            <select 
              value={reason} 
              onChange={(e) => setReason(e.target.value)}
              required
            >
              <option value="">Select a reason</option>
              {reportReasons.map(r => (
                <option key={r.value} value={r.value}>
                  {r.label}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Description (Optional)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Please provide additional details..."
              rows={4}
              maxLength={2000}
            />
            <small>{description.length}/2000 characters</small>
          </div>

          {error && <div className="error">{error}</div>}

          <div className="actions">
            <button type="submit" disabled={loading || !reason}>
              {loading ? 'Submitting...' : 'Submit Report'}
            </button>
            <button type="button" onClick={onClose}>
              Cancel
            </button>
          </div>
        </form>

        <div className="info">
          <p><strong>Note:</strong> False reports may result in account suspension.</p>
        </div>
      </div>
    </div>
  );
};
```

---

## Security Considerations

### 1. Anti-Abuse Measures

**Duplicate Prevention:**
```go
// Prevent reporting same user within 24 hours
var existingCount int
err := db.QueryRow(`
    SELECT COUNT(*) FROM reports 
    WHERE reporter_id = $1 
    AND reported_user_id = $2 
    AND created_at > NOW() - INTERVAL '24 hours'
`, reporterID, reportedUserID).Scan(&existingCount)

if existingCount > 0 {
    return errors.New("duplicate report within 24 hours")
}
```

**Self-Report Prevention:**
- Database constraint: `CHECK (reporter_id != reported_user_id)`
- Prevents users from reporting themselves

### 2. Privacy Protection

- Reporter identity hidden from reported user
- Admin notes not visible to regular users
- Sensitive descriptions not logged publicly

### 3. Admin Authorization

- All admin endpoints require `adminAuthMiddleware`
- Checks `is_admin` flag in database
- Returns 403 Forbidden for non-admin users

---

## Best Practices

### For Admins

1. **Response Time:** Review pending reports within 24 hours
2. **Documentation:** Always add detailed admin notes
3. **Consistency:** Use similar wording for similar violations
4. **Escalation:** Flag severe violations (violence threats, underage) immediately
5. **Pattern Recognition:** Track repeat offenders across multiple reports

### For Users

1. **Be Specific:** Provide detailed descriptions with evidence
2. **One Report Per Issue:** Don't spam multiple reports for same issue
3. **Evidence:** Include dates, times, specific messages/content
4. **Follow-Up:** Check status in "My Reports" section

---

## Report Statistics

### Admin Dashboard Queries

**Get report counts by status:**
```sql
SELECT 
    status,
    COUNT(*) as count
FROM reports
GROUP BY status;
```

**Get most reported users:**
```sql
SELECT 
    reported_user_id,
    u.username,
    COUNT(*) as report_count
FROM reports r
JOIN users u ON r.reported_user_id = u.user_id
GROUP BY reported_user_id, u.username
ORDER BY report_count DESC
LIMIT 10;
```

**Get reports by reason:**
```sql
SELECT 
    reason,
    COUNT(*) as count
FROM reports
GROUP BY reason
ORDER BY count DESC;
```

**Get average resolution time:**
```sql
SELECT 
    AVG(updated_at - created_at) as avg_resolution_time
FROM reports
WHERE status IN ('resolved', 'dismissed');
```

---

## Testing

### Manual Testing

**1. Create a report:**
```bash
curl -X POST http://localhost:8080/reports \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reported_user_id": 42,
    "reason": "harassment",
    "description": "User sent threatening messages"
  }'
```

**2. Get my reports:**
```bash
curl -X GET http://localhost:8080/reports/my \
  -H "Authorization: Bearer USER_TOKEN"
```

**3. Admin - Get all reports:**
```bash
curl -X GET "http://localhost:8080/admin/reports?status=pending" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**4. Admin - Update report:**
```bash
curl -X PATCH http://localhost:8080/admin/reports/1 \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved",
    "admin_notes": "User warned and content removed"
  }'
```

---

## Troubleshooting

### "Duplicate report" Error

**Symptom:** Cannot report a user.

**Cause:** Already reported within last 24 hours.

**Solution:** Wait 24 hours or contact admin if urgent.

### "Cannot report yourself" Error

**Symptom:** Report fails when testing.

**Cause:** Trying to report own account.

**Solution:** Use different user accounts for testing.

### Reports Not Appearing in Admin Panel

**Symptom:** Reports created but not visible.

**Solutions:**
1. Verify admin authentication token
2. Check status filter (reports may be in different status)
3. Check database directly: `SELECT * FROM reports;`

---

## Future Enhancements

- [ ] Report analytics dashboard
- [ ] Automated content moderation (AI-powered)
- [ ] Report appeal system
- [ ] User reputation scoring
- [ ] Automatic actions for repeat offenders
- [ ] Report categories for specific content types (photos, messages, profiles)
- [ ] Integration with chat system to attach message evidence
- [ ] Email notifications for report status changes

---

**Last Updated:** November 13, 2025
