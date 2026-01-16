/**
 * Content Sanitizer
 * Protects against prompt injection attacks
 */

const SUSPICIOUS_PATTERNS = [
  /ignore\s+(previous|all)\s+(instructions?|commands?)/i,
  /forget\s+(everything|all)/i,
  /you\s+are\s+now/i,
  /system\s*:\s*/i,
  /<\|(system|user|assistant)\|>/i,
  /\[INST\]/i,
  /\[SYSTEM\]/i,
  /override/i,
  /bypass/i,
  /ignore\s+the\s+above/i
];

const MAX_CONTENT_LENGTH = 1000000; // 1MB limit

/**
 * Sanitize content before sending to AI
 */
export function sanitizeContent(content) {
  const sanitized = {};

  // Handle text content
  if (content.text) {
    sanitized.text = sanitizeText(content.text);
  }

  // Handle DOM structure
  if (content.dom) {
    sanitized.dom = sanitizeDOM(content.dom);
  }

  // Handle accessibility tree
  if (content.accessibility) {
    sanitized.accessibility = sanitizeAccessibility(content.accessibility);
  }

  return sanitized;
}

/**
 * Sanitize plain text
 */
function sanitizeText(text) {
  if (!text || typeof text !== 'string') {
    return '';
  }

  // Check length
  if (text.length > MAX_CONTENT_LENGTH) {
    text = text.substring(0, MAX_CONTENT_LENGTH) + '...[truncated]';
  }

  // Check for suspicious patterns
  const hasSuspiciousPattern = SUSPICIOUS_PATTERNS.some(pattern => pattern.test(text));
  
  if (hasSuspiciousPattern) {
    // Flag but don't block - log warning
    console.warn('[Sanitizer] Suspicious pattern detected in text content');
    // Could add metadata flag here
  }

  return text;
}

/**
 * Sanitize DOM structure
 */
function sanitizeDOM(dom) {
  if (!dom || typeof dom !== 'object') {
    return null;
  }

  // Recursively sanitize node text content
  const sanitizeNode = (node) => {
    if (!node) return null;

    const sanitized = { ...node };

    // Sanitize text content
    if (sanitized.nodeValue) {
      sanitized.nodeValue = sanitizeText(sanitized.nodeValue);
    }

    // Sanitize attributes
    if (sanitized.attributes) {
      sanitized.attributes = sanitized.attributes.map(attr => ({
        ...attr,
        value: sanitizeText(attr.value || '')
      }));
    }

    // Recursively sanitize children
    if (sanitized.children) {
      sanitized.children = sanitized.children.map(sanitizeNode).filter(Boolean);
    }

    return sanitized;
  };

  return sanitizeNode(dom);
}

/**
 * Sanitize accessibility tree
 */
function sanitizeAccessibility(accessibility) {
  if (!accessibility || typeof accessibility !== 'object') {
    return null;
  }

  const sanitizeNode = (node) => {
    if (!node) return null;

    const sanitized = { ...node };

    // Sanitize text properties
    if (sanitized.name) {
      sanitized.name = sanitizeText(sanitized.name);
    }

    if (sanitized.value) {
      sanitized.value = sanitizeText(sanitized.value);
    }

    if (sanitized.description) {
      sanitized.description = sanitizeText(sanitized.description);
    }

    // Recursively sanitize children
    if (sanitized.children) {
      sanitized.children = sanitized.children.map(sanitizeNode).filter(Boolean);
    }

    return sanitized;
  };

  // Handle both single node and nodes array
  if (accessibility.nodes) {
    return {
      ...accessibility,
      nodes: accessibility.nodes.map(sanitizeNode).filter(Boolean)
    };
  }

  return sanitizeNode(accessibility);
}

/**
 * Check if content contains suspicious patterns
 */
export function hasSuspiciousContent(content) {
  const text = typeof content === 'string' ? content : JSON.stringify(content);
  return SUSPICIOUS_PATTERNS.some(pattern => pattern.test(text));
}
