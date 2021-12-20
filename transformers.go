package main

// *** REQUEST TRANSFORMERS ***

type IRequestTransformer interface {
	Transform(req *APIWrapper) (*APIWrapper, error)
	ShouldExpandRequest() bool
}

type RequestTransformers struct {
	transformers []IRequestTransformer
}

func (t *RequestTransformers) ShouldExpandRequest() bool {
	expand := false
	for _, tx := range t.transformers {
		if tx.ShouldExpandRequest() {
			expand = true
		}
	}
	return expand
}

func (t *RequestTransformers) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	defer func() {

	}()
	for _, transformer := range t.transformers {
		if _, err := transformer.Transform(wrapper); err != nil {
			return wrapper, err
		}
	}
	return wrapper, nil
}

func (t *RequestTransformers) Push(transformer IRequestTransformer) {
	t.transformers = append(t.transformers, transformer)
}

func NewRequestTransformers(transformers *[]TransformerConfig) (*RequestTransformers, error) {
	res := RequestTransformers{}
	for _, t := range *transformers {
		switch t.Id {
		case "url":
			transformer, err := NewRequestUrlTransformerFromParams(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "headers":
			transformer, err := NewRequestHeadersTransformerFromParams(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "basicAuth":
			transformer, err := NewBasicAuthTransformer(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "jwtAuth":
			transformer, err := NewJWTAuthTransformer(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "scriptable":
			transformer, err := NewScriptableTransformer(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "delay":
			transformer, _ := NewDelayTransformer(t.Params)
			res.Push(transformer)
		case "barrage":
			transformer, _ := NewBarrageRequestTransformer(t.Params)
			res.Push(transformer)
		}
	}
	return &res, nil
}

// *** RESPONSE TRANSFORMERS ***

type IResponseTransformer interface {
	Transform(wrapper *APIWrapper) (*APIWrapper, error)
	ShouldExpandRequest() bool
	ShouldExpandResponse() bool
}

type ResponseTransformers struct {
	transformers []IResponseTransformer
}

func (t *ResponseTransformers) ShouldExpandRequest() bool {
	expand := false
	for _, tx := range t.transformers {
		if tx.ShouldExpandRequest() {
			expand = true
		}
	}
	return expand
}

func (t *ResponseTransformers) ShouldExpandResponse() bool {
	expand := false
	for _, tx := range t.transformers {
		if tx.ShouldExpandResponse() {
			expand = true
		}
	}
	return expand
}

func (t *ResponseTransformers) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for _, transformer := range t.transformers {
		if wrapper, err := transformer.Transform(wrapper); err != nil {
			return wrapper, err
		}
	}
	return wrapper, nil
}

func (t *ResponseTransformers) Push(transformer IResponseTransformer) {
	t.transformers = append(t.transformers, transformer)
}

func NewResponseTransformers(transformers *[]TransformerConfig) (*ResponseTransformers, error) {
	res := ResponseTransformers{}
	for _, t := range *transformers {
		switch t.Id {
		case "headers":
			transformer, err := NewResponseHeadersTransformerFromParams(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "scriptable":
			transformer, err := NewScriptableTransformer(t.Params)
			if err != nil {
				return nil, err
			}
			res.Push(transformer)
		case "delay":
			transformer, _ := NewDelayTransformer(t.Params)
			res.Push(transformer)
		}
	}
	return &res, nil
}
