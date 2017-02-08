package net.isucon6.qualify.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class Paging {
    private int lastPage;
    private int[] pages;
}
