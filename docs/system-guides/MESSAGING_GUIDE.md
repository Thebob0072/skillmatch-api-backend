# 💬 Messaging System Guide

## Overview

The SkillMatch platform includes a **real-time messaging system** that allows clients and providers to communicate directly. The system features:

- Real-time chat via WebSocket
- Conversation management
- Message history
- Read receipts
- Typing indicators
- Message deletion

---

## Database Schema

### `conversations` Table

Stores chat conversations between two users.

```sql
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    user1_id INTEGER REFERENCES users(user_id),
    user2_id INTEGER REFERENCES users(user_id),
    last_message_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_user_order CHECK (user1_id < user2_id),
    CONSTRAINT unique_conversation UNIQUE (user1_id, user2_id)
);
```

**Key Points:**
- `user1_id` is always less than `user2_id` to prevent duplicates
- Automatically tracks last message timestamp
- Each user pair can only have one conversation

### `messages` Table

Stores individual messages within conversations.

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id),
    sender_id INTEGER REFERENCES users(user_id),
    receiver_id INTEGER REFERENCES users(user_id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',  -- 'text', 'image', 'system'
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Key Points:**
- Supports text, image, and system messages
- Tracks read status and read timestamp
- Content limited to 5000 characters

---

## API Endpoints

### 1. Get All Conversations

**GET** `/conversations`

Returns all conversations for the authenticated user.

**Response:**
```json
{
  "conversations": [
    {
      "id": 1,
      "user1_id": 2,
      "user2_id": 5,
      "last_message_at": "2025-11-13T10:30:00Z",
      "created_at": "2025-11-10T08:00:00Z",
      "updated_at": "2025-11-13T10:30:00Z",
      "other_user": {
        "id": 5,
        "display_name": "Jane Provider",
        "profile_photo_url": "https://...",
        "is_online": true
      },
      "last_message": {
        "content": "Thank you for your message!",
        "created_at": "2025-11-13T10:30:00Z"
      },
      "unread_count": 2
    }
  ],
  "total": 1,
  "unread_total": 2
}
```

---

### 2. Get Messages in a Conversation

**GET** `/conversations/:id/messages?limit=50&offset=0`

Returns all messages in a specific conversation.

**Query Parameters:**
- `limit` (optional): Number of messages to return (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:**
```json
{
  "messages": [
    {
      "id": 10,
      "conversation_id": 1,
      "sender_id": 5,
      "receiver_id": 2,
      "content": "Hello! I'm interested in your services.",
      "message_type": "text",
      "is_read": true,
      "read_at": "2025-11-13T10:15:00Z",
      "created_at": "2025-11-13T10:00:00Z",
      "sender_name": "Jane Provider",
      "sender_photo_url": "https://..."
    }
  ],
  "total": 1,
  "conversation_id": 1,
  "has_more": false
}
```

---

### 3. Send a Message

**POST** `/messages`

Sends a new message. Creates a conversation if it doesn't exist.

**Request Body:**
```json
{
  "receiver_id": 5,
  "content": "Hello! I'd like to book your service.",
  "message_type": "text"  // optional, defaults to "text"
}
```

**Response:**
```json
{
  "id": 11,
  "conversation_id": 1,
  "sender_id": 2,
  "receiver_id": 5,
  "content": "Hello! I'd like to book your service.",
  "message_type": "text",
  "is_read": false,
  "created_at": "2025-11-13T11:00:00Z"
}
```

**Side Effects:**
- Creates a new conversation if one doesn't exist
- Sends real-time notification via WebSocket to receiver
- Creates a notification record

---

### 4. Mark Messages as Read

**PATCH** `/messages/read`

Marks one or more messages as read.

**Request Body:**
```json
{
  "message_ids": [10, 11, 12]
}
```

**Response:**
```json
{
  "updated_count": 3,
  "message_ids": [10, 11, 12]
}
```

**Side Effects:**
- Sends read receipts via WebSocket to senders

---

### 5. Delete a Message

**DELETE** `/messages/:id`

Soft deletes a message (replaces content with "[Message deleted]").

**Response:**
```json
{
  "message": "Message deleted"
}
```

**Rules:**
- Only the sender can delete their own messages
- Content is replaced with "[Message deleted]"
- Message type changed to "system"

---

## WebSocket Connection

### Connecting to WebSocket

**Endpoint:** `ws://localhost:8080/ws`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Connection Example (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

ws.onopen = () => {
  console.log('Connected to WebSocket');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleWebSocketMessage(data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

---

### WebSocket Message Types

#### 1. New Message

**Received when someone sends you a message:**

```json
{
  "type": "message",
  "payload": {
    "id": 15,
    "conversation_id": 2,
    "sender_id": 3,
    "receiver_id": 1,
    "content": "Hey there!",
    "message_type": "text",
    "is_read": false,
    "created_at": "2025-11-13T12:00:00Z"
  }
}
```

#### 2. Read Receipt

**Received when someone reads your message:**

```json
{
  "type": "read_receipt",
  "payload": {
    "conversation_id": 2,
    "message_ids": [14, 15]
  }
}
```

#### 3. Typing Indicator

**Send this to indicate you're typing:**

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 2,
    "is_typing": true
  }
}
```

**Receive when someone is typing:**

```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 2,
    "user_id": 3,
    "is_typing": true
  }
}
```

#### 4. Ping/Pong (Keep-Alive)

**Send:**
```json
{
  "type": "ping"
}
```

**Receive:**
```json
{
  "type": "pong",
  "payload": {
    "timestamp": 1699876543
  }
}
```

---

## Frontend Implementation Example

### React Component Example

```typescript
import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';

interface Message {
  id: number;
  sender_id: number;
  receiver_id: number;
  content: string;
  created_at: string;
  sender_name: string;
}

interface ChatProps {
  conversationId: number;
  currentUserId: number;
}

export const Chat: React.FC<ChatProps> = ({ conversationId, currentUserId }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Connect to WebSocket
  useEffect(() => {
    const token = localStorage.getItem('token');
    const ws = new WebSocket(`ws://localhost:8080/ws`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'message' && data.payload.conversation_id === conversationId) {
        setMessages(prev => [...prev, data.payload]);
        
        // Mark as read
        axios.patch('/messages/read', {
          message_ids: [data.payload.id]
        });
      }
      
      if (data.type === 'typing' && data.payload.conversation_id === conversationId) {
        setIsTyping(data.payload.is_typing);
      }
    };

    wsRef.current = ws;

    return () => ws.close();
  }, [conversationId]);

  // Load messages
  useEffect(() => {
    axios.get(`/conversations/${conversationId}/messages`)
      .then(res => setMessages(res.data.messages));
  }, [conversationId]);

  // Send message
  const sendMessage = async () => {
    if (!newMessage.trim()) return;

    const response = await axios.post('/messages', {
      receiver_id: /* other user id */,
      content: newMessage
    });

    setMessages(prev => [...prev, response.data]);
    setNewMessage('');
  };

  // Handle typing indicator
  const handleTyping = () => {
    wsRef.current?.send(JSON.stringify({
      type: 'typing',
      payload: {
        conversation_id: conversationId,
        is_typing: true
      }
    }));

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    typingTimeoutRef.current = setTimeout(() => {
      wsRef.current?.send(JSON.stringify({
        type: 'typing',
        payload: {
          conversation_id: conversationId,
          is_typing: false
        }
      }));
    }, 2000);
  };

  return (
    <div className="chat-container">
      <div className="messages">
        {messages.map(msg => (
          <div key={msg.id} className={msg.sender_id === currentUserId ? 'my-message' : 'their-message'}>
            <p>{msg.content}</p>
            <small>{new Date(msg.created_at).toLocaleTimeString()}</small>
          </div>
        ))}
        {isTyping && <p className="typing-indicator">Typing...</p>}
      </div>
      
      <div className="input-area">
        <input
          value={newMessage}
          onChange={(e) => {
            setNewMessage(e.target.value);
            handleTyping();
          }}
          onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="Type a message..."
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
};
```

---

## Security Considerations

### 1. Authorization
- Users can only see conversations they're part of
- Users can only send messages to verified users
- WebSocket connections require valid JWT token

### 2. Rate Limiting
- Implement rate limiting on message sending (e.g., 10 messages per minute)
- Prevent spam and abuse

### 3. Content Moderation
- Validate message content length (max 5000 characters)
- Consider adding profanity filters
- Allow users to report inappropriate messages

### 4. Data Privacy
- Messages are only visible to conversation participants
- Consider implementing end-to-end encryption for sensitive platforms
- Deleted messages are soft-deleted (content replaced)

---

## Performance Optimization

### 1. Pagination
- Load messages in batches (default: 50)
- Implement infinite scroll for message history

### 2. WebSocket Connection Management
- Auto-reconnect on connection loss
- Send ping/pong every 30 seconds to keep connection alive
- Close old connections when user opens multiple tabs

### 3. Caching
- Cache conversation list on client
- Update cache on new messages
- Invalidate cache on logout

### 4. Database Indexes
- Index on `conversation_id` for fast message lookup
- Index on `receiver_id` + `is_read` for unread count queries
- Index on `created_at` for chronological ordering

---

## Testing

### Manual Testing with cURL

**1. Send a message:**
```bash
curl -X POST http://localhost:8080/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_id": 5,
    "content": "Hello, this is a test message!"
  }'
```

**2. Get conversations:**
```bash
curl -X GET http://localhost:8080/conversations \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**3. Mark messages as read:**
```bash
curl -X PATCH http://localhost:8080/messages/read \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message_ids": [1, 2, 3]
  }'
```

### WebSocket Testing (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('✅ Connected');
  
  // Send a typing indicator
  ws.send(JSON.stringify({
    type: 'typing',
    payload: { conversation_id: 1, is_typing: true }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('📩 Received:', data);
};
```

---

## Troubleshooting

### WebSocket Connection Fails

**Symptom:** WebSocket connection immediately closes or never opens.

**Solutions:**
1. Check JWT token is valid and not expired
2. Verify `Authorization` header is included
3. Check firewall/proxy allows WebSocket connections
4. Ensure server is running and `/ws` endpoint is registered

### Messages Not Appearing in Real-Time

**Symptom:** Messages only appear after page refresh.

**Solutions:**
1. Verify WebSocket connection is established
2. Check `conversation_id` matches in WebSocket handler
3. Ensure `wsManager.BroadcastToUser()` is called after sending message
4. Check browser console for WebSocket errors

### "User not found" Error

**Symptom:** Cannot send message to a user.

**Solutions:**
1. Verify `receiver_id` is valid
2. Check receiver user exists in database
3. Ensure receiver has completed KYC verification (if required)

---

## Future Enhancements

- [ ] End-to-end encryption
- [ ] File/image attachments
- [ ] Voice/video messages
- [ ] Message search
- [ ] Message reactions (like, heart, etc.)
- [ ] Group chat support
- [ ] Message forwarding
- [ ] Auto-delete messages after X days
- [ ] Mute conversations
- [ ] Block users

---

**Last Updated:** November 13, 2025
