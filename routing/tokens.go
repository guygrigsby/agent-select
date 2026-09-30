package routing

// Closed vocabularies from educated-routing.md. Every list is exhaustive;
// unknown tokens are invalid at construction.

var taskTokens = map[string]bool{
	"general": true, "code": true, "review": true, "reasoning": true,
	"extraction": true, "writing": true, "translation": true, "vision": true,
	"image_generation": true, "image_editing": true,
}

var modeTokens = map[string]bool{
	"cost": true, "performance": true,
}

var capabilityTokens = map[string]bool{
	"text_input": true, "text_output": true, "image_input": true,
	"image_output": true, "tool_use": true, "structured_output": true,
	"image_edit": true, "image_reference": true, "image_mask": true,
	"image_alpha": true,
}

var resourceTokens = map[string]bool{
	"input_tokens": true, "output_tokens": true, "context_tokens": true,
	"image_width": true, "image_height": true, "image_bytes": true,
	"input_images": true, "output_images": true, "references": true,
}

var priceComponentTokens = map[string]bool{
	"input_tokens": true, "output_tokens": true, "input_images": true,
	"output_images": true, "references": true, "searches": true,
}

var basisTokens = map[string]bool{
	"published": true, "publisher_claim": true, "host_observed": true,
}

var imageOperationTokens = map[string]bool{
	"none": true, "vision": true, "generate": true, "edit": true,
}

var distributionTokens = map[string]bool{
	"local": true, "distributable": true,
}
