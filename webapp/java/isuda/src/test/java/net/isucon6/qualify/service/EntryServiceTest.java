package net.isucon6.qualify.service;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Date;

import net.isucon6.qualify.domain.Entry;
import net.isucon6.qualify.mapper.EntryMapper;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.runners.MockitoJUnitRunner;
import org.modelmapper.ModelMapper;
import org.springframework.beans.factory.annotation.Autowired;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.is;

@RunWith(MockitoJUnitRunner.class)
public class EntryServiceTest {
    @Mock
    private EntryMapper entryMapper;
    @Autowired
    private ModelMapper modelMapper;

    @Test
    public void testHtmlify() throws NoSuchMethodException, InvocationTargetException, IllegalAccessException {
        String description = "123の説明です. 以下の文字列はリンクのはずです. ジゴロ / re[ge]xp / /html/escape";
        Mockito.when(entryMapper.findAllOrderByLength()).thenReturn(new ArrayList<Entry>() {{
            add(new Entry(1L, 1L, "123", description, new Date(), new Date()));
            add(new Entry(2L, 1L, "ジゴロ", "日本語文字列の置換の検査", new Date(), new Date()));
            add(new Entry(3L, 1L, "re[ge]xp", "正規表現が入っててもエスケープされる検査", new Date(), new Date()));
            add(new Entry(4L, 1L, "/html/escape", "HTML特殊文字がエスケープされる検査", new Date(), new Date()));
        }});

        EntryService entryService = new EntryService(entryMapper, modelMapper);
        Method method = entryService.getClass().getDeclaredMethod("htmlify", String.class);
        method.setAccessible(true);
        String actual = (String) method.invoke(entryService, description);

        assertThat(actual, is("<a href=\"/keyword/123\">123</a>の説明です. 以下の文字列はリンクのはずです. <a href=\"/keyword/%E3%82%B8%E3%82%B4%E3%83%AD\">ジゴロ</a> / <a href=\"/keyword/re%5Bge%5Dxp\">re[ge]xp</a> / <a href=\"/keyword/%2Fhtml%2Fescape\">/html/escape</a>"));
    }
}
