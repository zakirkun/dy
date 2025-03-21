package dy

// ContextField represents a key-value pair in the logging context
type ContextField struct {
	Key   string
	Value interface{}
}

// LogContext contains all contextual fields for a logger instance
type LogContext struct {
	Fields []ContextField
}

// Add adds a new field to the context
func (c *LogContext) Add(key string, value interface{}) {
	c.Fields = append(c.Fields, ContextField{Key: key, Value: value})
}

// Clone creates a copy of the context
func (c *LogContext) Clone() *LogContext {
	if c == nil {
		return &LogContext{}
	}

	clone := &LogContext{
		Fields: make([]ContextField, len(c.Fields)),
	}

	copy(clone.Fields, c.Fields)
	return clone
}

// Remove removes a field from the context by key
func (c *LogContext) Remove(key string) {
	if c == nil {
		return
	}

	for i, field := range c.Fields {
		if field.Key == key {
			// Remove the item by swapping with the last element and slicing
			c.Fields[i] = c.Fields[len(c.Fields)-1]
			c.Fields = c.Fields[:len(c.Fields)-1]
			return
		}
	}
}

// WithContext creates a new logger with additional context fields
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a new logger that shares the same configuration
	child := &Logger{
		out:          l.out,
		level:        l.level,
		prefix:       l.prefix,
		timestamp:    l.timestamp,
		nestingLevel: l.nestingLevel,
		traceEnabled: l.traceEnabled,
		indentString: l.indentString,
		jsonFormat:   l.jsonFormat,
		callerInfo:   l.callerInfo,
		colorEnabled: l.colorEnabled,
		closer:       l.closer,
	}

	// Clone the context if it exists, or create a new one
	child.context = l.context.Clone()
	if child.context == nil {
		child.context = &LogContext{}
	}

	// Add the new field to the context
	child.context.Add(key, value)

	return child
}

// WithFields creates a new logger with multiple additional context fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a new logger that shares the same configuration
	child := &Logger{
		out:          l.out,
		level:        l.level,
		prefix:       l.prefix,
		timestamp:    l.timestamp,
		nestingLevel: l.nestingLevel,
		traceEnabled: l.traceEnabled,
		indentString: l.indentString,
		jsonFormat:   l.jsonFormat,
		callerInfo:   l.callerInfo,
		colorEnabled: l.colorEnabled,
		closer:       l.closer,
	}

	// Clone the context if it exists, or create a new one
	child.context = l.context.Clone()
	if child.context == nil {
		child.context = &LogContext{}
	}

	// Add all the new fields to the context
	for k, v := range fields {
		child.context.Add(k, v)
	}

	return child
}

// WithoutContext creates a new logger without the specified context key
func (l *Logger) WithoutContext(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a new logger that shares the same configuration
	child := &Logger{
		out:          l.out,
		level:        l.level,
		prefix:       l.prefix,
		timestamp:    l.timestamp,
		nestingLevel: l.nestingLevel,
		traceEnabled: l.traceEnabled,
		indentString: l.indentString,
		jsonFormat:   l.jsonFormat,
		callerInfo:   l.callerInfo,
		colorEnabled: l.colorEnabled,
		closer:       l.closer,
	}

	// Clone the context if it exists
	child.context = l.context.Clone()
	if child.context != nil {
		child.context.Remove(key)
	}

	return child
}
