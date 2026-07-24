package main

const (
	systemPrompt = `You are an AI assistant for a creative writing application. Your goal is to provide helpful and inspiring suggestions to the writer. You must always respond with valid JSON exactly matching the requested format. Do not include markdown code fences or additional text.`

	wordErrorPrompt = `The following is a paragraph from a story. Please identify any misspelled or grammatically incorrect words. For each error, provide the incorrect word and a suggested correction. Please format your response as a JSON array of objects, where each object has "incorrect" and "correct" keys. If there are no errors, return an empty array [].`

	themePrompt = `The following is a paragraph from a story. Based on the mood and tone of the paragraph, please suggest a visual theme for the writing application. The available themes are "midnight", "parchment", "blossom", and "studio". Please format your response as a JSON object with a "theme" key. Pick exactly one of the four options.`

	fontPrompt = `The following is a paragraph from a story. Based on the mood and tone of the paragraph, please suggest a font for the writing application. The available fonts are "literary", "editorial", and "typewriter". Please format your response as a JSON object with a "font" key. Pick exactly one of the three options.`

	illustrationPrompt = `The following is a paragraph from a story. Choose 2-3 simple keywords (only from: nature, landscape, forest, ocean, mountain, desert, city, street, room, window, rain, snow, fog, mist, storm, sunrise, sunset, night, dawn, autumn, spring, summer, winter, cozy, dark, bright, moody, minimal, vintage, modern, warm, cold, quiet, lonely, peaceful) that best match the scene. Format as JSON: {"illustration": "keyword1 keyword2"}. Return ONLY keywords from this list separated by spaces — no other text.`

	wordAutocompletePrompt = `The writer is currently typing and has started the word "%s". Based on the surrounding text, suggest 3 to 5 complete words that could be what the writer intends. Consider context, mood, and literary quality. Please format your response as a JSON object with a "words" key containing an array of strings. Return only plausible, complete words — not phrases.`

	paragraphAutocompletePrompt = `The following is text from a story that ends mid-thought. Please suggest a natural continuation — a single sentence (or at most two short sentences) that flows smoothly from where the writer left off. Match the tone, style, and voice of the existing text. Please format your response as a JSON object with a "continuation" key containing the suggested sentence. Do not repeat the existing text.`

	fullParagraphAutocompletePrompt = `The following is text from a story ending mid-thought. Based on the tone, style, and narrative direction, write a complete new paragraph that continues the story naturally. Match the existing voice and pacing. The paragraph should be 3-6 sentences. Please format your response as a JSON object with a "continuation" key containing the full paragraph. Do not repeat the existing text.`
)