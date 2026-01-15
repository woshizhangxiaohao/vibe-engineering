"use client";

import { useState, useEffect, useCallback, RefObject } from "react";

/**
 * Text selection information
 */
export interface TextSelection {
  text: string;
  startOffset: number;
  endOffset: number;
  rect: DOMRect;
}

/**
 * Hook for detecting text selection within a container
 * @param containerRef - Reference to the container element
 * @returns Selection state and clear function
 */
export function useTextSelection(containerRef: RefObject<HTMLElement | null>) {
  const [selection, setSelection] = useState<TextSelection | null>(null);

  const clearSelection = useCallback(() => {
    setSelection(null);
    window.getSelection()?.removeAllRanges();
  }, []);

  const handleMouseUp = useCallback(() => {
    if (!containerRef.current) return;

    const windowSelection = window.getSelection();
    if (!windowSelection || windowSelection.rangeCount === 0) {
      clearSelection();
      return;
    }

    const range = windowSelection.getRangeAt(0);
    const selectedText = windowSelection.toString().trim();

    // Only proceed if there's selected text
    if (!selectedText) {
      clearSelection();
      return;
    }

    // Check if selection is within our container
    const container = containerRef.current;
    if (!container.contains(range.commonAncestorContainer)) {
      clearSelection();
      return;
    }

    // Get the bounding rect for positioning the toolbar
    const rect = range.getBoundingClientRect();

    // Calculate offsets relative to the container's text content
    const startOffset = getTextOffset(container, range.startContainer, range.startOffset);
    const endOffset = getTextOffset(container, range.endContainer, range.endOffset);

    setSelection({
      text: selectedText,
      startOffset,
      endOffset,
      rect,
    });
  }, [containerRef, clearSelection]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // Listen to mouseup events for text selection
    container.addEventListener("mouseup", handleMouseUp);

    // Clear selection when clicking outside
    const handleClickOutside = (e: MouseEvent) => {
      if (container && !container.contains(e.target as Node)) {
        clearSelection();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      container.removeEventListener("mouseup", handleMouseUp);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [containerRef, handleMouseUp, clearSelection]);

  return { selection, clearSelection };
}

/**
 * Calculate text offset from the start of the container
 * @param container - Root container element
 * @param node - Current text node
 * @param offset - Offset within the current node
 * @returns Total offset from container start
 */
function getTextOffset(
  container: Node,
  node: Node,
  offset: number
): number {
  let textOffset = 0;
  const walker = document.createTreeWalker(
    container,
    NodeFilter.SHOW_TEXT,
    null
  );

  let currentNode = walker.nextNode();
  while (currentNode) {
    if (currentNode === node) {
      return textOffset + offset;
    }
    textOffset += currentNode.textContent?.length || 0;
    currentNode = walker.nextNode();
  }

  return textOffset;
}
