package net.isucon6.qualify.service;

import java.util.stream.IntStream;

import net.isucon6.qualify.dto.Paging;
import net.isucon6.qualify.mapper.EntryMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class PagingService {
    @Autowired
    private EntryMapper entryMapper;

    public Paging fetchEntryPage(double perPage, double page) {
        double count = entryMapper.fetchCount();
        int lastPage = (int) Math.ceil(count / perPage);

        int[] pages = IntStream.range((int) Math.max(1, page - 5), (int) Math.min(lastPage, page + 5) + 1).toArray();
        return new Paging(lastPage, pages);
    }
}
