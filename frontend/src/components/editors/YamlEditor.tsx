import Editor, { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/editor/editor.worker?worker';

// Bundle Monaco locally (no CDN) so the editor works offline / behind NAT.
// Configure the worker so Monaco does not fall back to synchronous parsing.
self.MonacoEnvironment = {
    getWorker: () => new editorWorker(),
};

loader.config({ monaco });

interface YamlEditorProps {
    value: string;
    onChange?: (value: string) => void;
    height?: number;
    readOnly?: boolean;
}

/**
 * Lightweight Monaco-based YAML editor used by the Compose tab.
 * Wrapped in a component so a richer editor (schema validation, linting) can
 * be swapped in later without touching the page code.
 */
export function YamlEditor({ value, onChange, height = 360, readOnly = false }: YamlEditorProps) {
    return (
        <div
            style={{
                border: '1px solid var(--color-border-2)',
                borderRadius: 4,
                overflow: 'hidden',
            }}
        >
            <Editor
                height={height}
                defaultLanguage="yaml"
                value={value}
                onChange={(v) => onChange?.(v ?? '')}
                options={{
                    readOnly,
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: 'on',
                    scrollBeyondLastLine: false,
                    tabSize: 2,
                    wordWrap: 'on',
                    automaticLayout: true,
                    formatOnPaste: true,
                }}
            />
        </div>
    );
}